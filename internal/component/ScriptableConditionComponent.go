package component

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
)

const ScriptableCondition ecs.ComponentType = "ScriptableConditionComponent"

// SpawnEntityFunc may be set by the factory package at startup to allow
// condition scripts to spawn entities without creating an import cycle.
// Signature: (blueprint string, x, y, z int, levelData any) error
var SpawnEntityFunc func(blueprint string, x, y, z int, levelData any) error

// conditionContext holds mutable state shared between the component and the
// closures registered on the interpreter. Level must be updated before each
// event call so that spawn_entity and similar functions always see the current
// level, even though the interpreter is only initialised once.
type conditionContext struct {
	Level any
}

// ScriptableConditionComponent is a decaying status condition driven by a
// MechBasic script. The script may define four event functions:
//
//	on_applied()  — called once when the condition is first attached
//	on_turn()     — called each turn (or every Interval turns if set)
//	on_death()    — called when the host entity dies while the condition is active
//	on_removed()  — called when the condition expires naturally
//
// Script functions available inside condition scripts:
//
//	get_x(), get_y(), get_z()                        — entity position
//	get_var(key), set_var(key, val)                  — condition-local variables
//	get_duration()                                   — remaining turns
//	add_message(msg)                                 — add to message log
//	deal_damage(amount, type)                        — deal damage to the entity
//	apply_damage_condition(name, dur, dice, type)    — add a DamageConditionComponent
//	spawn_entity(blueprint, x, y, z)                 — spawn an entity (requires level)
type ScriptableConditionComponent struct {
	Name       string         `json:"name,omitempty"`
	Duration   int            `json:"duration,omitempty"`
	ScriptPath string         `json:"script_path,omitempty"`
	Interval   int            `json:"interval,omitempty"` // call on_turn every N turns; 0/1 = every turn
	Vars       map[string]any `json:"vars,omitempty"`

	// runtime state — not serialized
	Interpreter *basic.MechBasic `json:"-"`
	ctx         *conditionContext
	turnCount   int
	applied     bool
}

func (c *ScriptableConditionComponent) GetType() ecs.ComponentType { return ScriptableCondition }
func (c *ScriptableConditionComponent) GetConditionName() string   { return c.Name }
func (c *ScriptableConditionComponent) getDuration() int           { return c.Duration }
func (c *ScriptableConditionComponent) setDuration(d int)          { c.Duration = d }

func (c *ScriptableConditionComponent) Decay() bool {
	c.Duration--
	return c.Duration <= 0
}

// ApplyOnce satisfies ConditionModifier. Called by ActiveConditionsComponent.Tick
// before damage/turn logic. Initializes the interpreter and fires on_applied once.
// Level is not available at this call site; spawn_entity will not work in on_applied.
func (c *ScriptableConditionComponent) ApplyOnce(entity *ecs.Entity) {
	if c.applied {
		return
	}
	c.applied = true
	if err := c.ensureInit(entity); err != nil {
		log.Printf("[ScriptableCondition] init failed for %q: %v", c.ScriptPath, err)
		return
	}
	c.callIfExists("on_applied")
}

// Revert satisfies ConditionModifier. Called by ActiveConditionsComponent.Tick
// when the condition expires (Decay returns true). Fires on_removed.
func (c *ScriptableConditionComponent) Revert(_ *ecs.Entity) {
	c.callIfExists("on_removed")
}

// OnDeath satisfies rlcomponents.DeathHandler. Called by FireDeath when the
// host entity dies. Fires the on_death script function if defined.
func (c *ScriptableConditionComponent) OnDeath(entity *ecs.Entity, levelData any) {
	if err := c.ensureInit(entity); err != nil {
		log.Printf("[ScriptableCondition] init failed for %q: %v", c.ScriptPath, err)
		return
	}
	c.ctx.Level = levelData
	c.callIfExists("on_death")
}

// HandleTurn satisfies rlcomponents.TurnHandler. Called by
// ActiveConditionsComponent.Tick each turn. Fires on_turn respecting Interval.
func (c *ScriptableConditionComponent) HandleTurn(entity *ecs.Entity, levelData any) {
	if err := c.ensureInit(entity); err != nil {
		log.Printf("[ScriptableCondition] init failed for %q: %v", c.ScriptPath, err)
		return
	}
	c.ctx.Level = levelData
	c.turnCount++
	interval := c.Interval
	if interval <= 0 {
		interval = 1
	}
	if c.turnCount%interval == 0 {
		c.callIfExists("on_turn")
	}
}

func (c *ScriptableConditionComponent) callIfExists(fn string) {
	if c.Interpreter == nil || !c.Interpreter.HasFunction(fn) {
		return
	}
	if _, err := c.Interpreter.Call(fn); err != nil {
		log.Printf("[ScriptableCondition] %s: %v", fn, err)
	}
}

func (c *ScriptableConditionComponent) ensureInit(entity *ecs.Entity) error {
	if c.Interpreter != nil {
		return nil
	}
	code, err := os.ReadFile(c.ScriptPath)
	if err != nil {
		return fmt.Errorf("read script %q: %w", c.ScriptPath, err)
	}

	c.ctx = &conditionContext{}
	interp := basic.NewMechanicalBasic()
	c.registerFuncs(interp, entity)

	if err := interp.Load(string(code)); err != nil {
		return fmt.Errorf("load script %q: %w", c.ScriptPath, err)
	}
	c.Interpreter = interp
	return nil
}

func (c *ScriptableConditionComponent) registerFuncs(interp *basic.MechBasic, entity *ecs.Entity) {
	interp.RegisterFunc("get_x", func(...any) (any, error) {
		pc := entity.GetComponent(Position).(*PositionComponent)
		return float64(pc.GetX()), nil
	})
	interp.RegisterFunc("get_y", func(...any) (any, error) {
		pc := entity.GetComponent(Position).(*PositionComponent)
		return float64(pc.GetY()), nil
	})
	interp.RegisterFunc("get_z", func(...any) (any, error) {
		pc := entity.GetComponent(Position).(*PositionComponent)
		return float64(pc.GetZ()), nil
	})

	interp.RegisterFunc("get_var", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("get_var: expected 1 argument")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("get_var: argument must be a string")
		}
		if c.Vars == nil {
			return nil, nil
		}
		return c.Vars[key], nil
	})
	interp.RegisterFunc("set_var", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set_var: expected 2 arguments")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("set_var: first argument must be a string")
		}
		if c.Vars == nil {
			c.Vars = make(map[string]any)
		}
		c.Vars[key] = args[1]
		return nil, nil
	})

	interp.RegisterFunc("get_duration", func(...any) (any, error) {
		return float64(c.Duration), nil
	})

	interp.RegisterFunc("add_message", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("add_message: expected 1 argument")
		}
		msg, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("add_message: argument must be a string")
		}
		message.AddMessage(msg)
		return nil, nil
	})

	interp.RegisterFunc("deal_damage", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("deal_damage: expected (amount, type)")
		}
		amt := condToInt(args[0])
		dtype, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("deal_damage: type argument must be a string")
		}
		applyScriptConditionDamage(entity, amt, dtype)
		return nil, nil
	})

	interp.RegisterFunc("apply_damage_condition", func(args ...any) (any, error) {
		if len(args) != 4 {
			return nil, fmt.Errorf("apply_damage_condition: expected (name, duration, dice, type)")
		}
		name, ok1 := args[0].(string)
		dur := condToInt(args[1])
		dice, ok2 := args[2].(string)
		dtype, ok3 := args[3].(string)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("apply_damage_condition: name, dice, and type must be strings")
		}
		acc := rlcomponents.GetOrCreateActiveConditions(entity)
		acc.Add(&rlcomponents.DamageConditionComponent{
			Name: name, Duration: dur, DamageDice: dice, DamageType: dtype,
		})
		return nil, nil
	})

	interp.RegisterFunc("spawn_entity", func(args ...any) (any, error) {
		if len(args) != 4 {
			return nil, fmt.Errorf("spawn_entity: expected (blueprint, x, y, z)")
		}
		if SpawnEntityFunc == nil {
			log.Printf("[ScriptableCondition] spawn_entity: no spawn function registered")
			return nil, nil
		}
		if c.ctx.Level == nil {
			log.Printf("[ScriptableCondition] spawn_entity: no level available")
			return nil, nil
		}
		blueprint, _ := args[0].(string)
		x, y, z := condToInt(args[1]), condToInt(args[2]), condToInt(args[3])
		return nil, SpawnEntityFunc(blueprint, x, y, z, c.ctx.Level)
	})
}

func condToInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// applyScriptConditionDamage deals dmg to a random non-amputated body part,
// then calls rlentity.HandleDeath to apply DeadComponent if the damage is
// lethal (KillsWhenBroken, KillsWhenAmputated, or Health <= 0).
// Falls back to HealthComponent when no BodyComponent is present.
// damageType is accepted for future use (resistance checks, etc.) but not yet applied.
func applyScriptConditionDamage(entity *ecs.Entity, dmg int, _ string) {
	if dmg <= 0 {
		return
	}
	if entity.HasComponent(Body) {
		bc := entity.GetComponent(Body).(*BodyComponent)
		var available []string
		for name, part := range bc.Parts {
			if !part.Amputated {
				available = append(available, name)
			}
		}
		if len(available) == 0 {
			entity.AddComponent(&rlcomponents.DeadComponent{})
			return
		}
		name := available[rand.Intn(len(available))]
		part := bc.Parts[name]
		part.HP -= dmg
		if part.HP <= 0 && !part.Broken {
			part.Broken = true
		}
		bc.Parts[name] = part
	} else if entity.HasComponent(rlcomponents.Health) {
		hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
		hc.Health -= dmg
		if hc.Health < 0 {
			hc.Health = 0
		}
	}
	rlentity.HandleDeath(entity)
}
