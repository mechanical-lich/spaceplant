package component

import (
	"fmt"
	"log"
	"os"

	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
)

const ScriptableCondition ecs.ComponentType = "ScriptableConditionComponent"

// SpawnEntityFunc may be set by the factory package at startup to allow
// condition scripts to spawn entities without creating an import cycle.
// Signature: (blueprint string, x, y, z int, levelData any) error
var SpawnEntityFunc func(blueprint string, x, y, z int, levelData any) error

// ScriptableConditionComponent is a decaying status condition driven by a
// MechBasic script. The script may define three event functions:
//
//	on_applied()  — called once when the condition is first attached
//	on_turn()     — called each turn (or every Interval turns if set)
//	on_removed()  — called when the condition expires or is removed
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
func (c *ScriptableConditionComponent) ApplyOnce(entity *ecs.Entity) {
	if c.applied {
		return
	}
	c.applied = true
	if err := c.ensureInit(entity, nil); err != nil {
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
	if err := c.ensureInit(entity, levelData); err != nil {
		log.Printf("[ScriptableCondition] init failed for %q: %v", c.ScriptPath, err)
		return
	}
	c.callIfExists("on_death")
}

// HandleTurn satisfies rlcomponents.TurnHandler. Called by
// ActiveConditionsComponent.Tick each turn. Fires on_turn respecting Interval.
func (c *ScriptableConditionComponent) HandleTurn(entity *ecs.Entity, levelData any) {
	if err := c.ensureInit(entity, levelData); err != nil {
		log.Printf("[ScriptableCondition] init failed for %q: %v", c.ScriptPath, err)
		return
	}
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

func (c *ScriptableConditionComponent) ensureInit(entity *ecs.Entity, levelData any) error {
	if c.Interpreter != nil {
		return nil
	}
	code, err := os.ReadFile(c.ScriptPath)
	if err != nil {
		return fmt.Errorf("read script %q: %w", c.ScriptPath, err)
	}

	interp := basic.NewMechanicalBasic()
	c.registerFuncs(interp, entity, levelData)

	if err := interp.Load(string(code)); err != nil {
		return fmt.Errorf("load script %q: %w", c.ScriptPath, err)
	}
	c.Interpreter = interp
	return nil
}

func (c *ScriptableConditionComponent) registerFuncs(interp *basic.MechBasic, entity *ecs.Entity, levelData any) {
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
		if len(args) < 1 {
			return nil, fmt.Errorf("deal_damage: expected amount argument")
		}
		amt := condToInt(args[0])
		applyScriptConditionDamage(entity, amt)
		return nil, nil
	})

	interp.RegisterFunc("apply_damage_condition", func(args ...any) (any, error) {
		if len(args) != 4 {
			return nil, fmt.Errorf("apply_damage_condition: expected (name, duration, dice, type)")
		}
		name, _ := args[0].(string)
		dur := condToInt(args[1])
		dice, _ := args[2].(string)
		dtype, _ := args[3].(string)
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
		if levelData == nil {
			log.Printf("[ScriptableCondition] spawn_entity: no level available")
			return nil, nil
		}
		blueprint, _ := args[0].(string)
		x, y, z := condToInt(args[1]), condToInt(args[2]), condToInt(args[3])
		return nil, SpawnEntityFunc(blueprint, x, y, z, levelData)
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

// applyScriptConditionDamage deals damage to the entity, mirroring the logic
// in rlsystems.applyStatusDamage (body part or health).
func applyScriptConditionDamage(entity *ecs.Entity, dmg int) {
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
		// Use first available part deterministically for now.
		name := available[0]
		part := bc.Parts[name]
		part.HP -= dmg
		if part.HP <= 0 && !part.Broken {
			part.Broken = true
		}
		bc.Parts[name] = part
	} else if entity.HasComponent(rlcomponents.Health) {
		hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
		hc.Health -= dmg
	}
}
