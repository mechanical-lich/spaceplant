package system

import (
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type ScriptSystem struct{}

func (s ScriptSystem) Requires() []ecs.ComponentType {
	return []ecs.ComponentType{component.Script}
}

func (s ScriptSystem) UpdateSystem(_ any) error { return nil }

func (s ScriptSystem) UpdateEntity(data any, entity *ecs.Entity) error {
	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}

	l, ok := data.(*world.Level)
	if !ok {
		return errors.New("ScriptSystem: invalid level type")
	}

	if err := ensureInit(entity, l); err != nil {
		return err
	}

	sc := entity.GetComponent(component.Script).(*component.ScriptComponent)
	sc.Ctx.(*scriptContext).Level = l

	if !entity.HasComponent(rlcomponents.MyTurn) {
		return nil
	}

	entity.AddComponent(rlcomponents.GetTurnTaken())
	if entity.HasComponent(rlcomponents.Energy) {
		ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		rlenergy.SetActionCost(entity, ec.Speed)
	}
	return callIfExists(sc, "on_turn_taken")
}

// CallScriptEvent initializes the script if needed and calls the named event
// function (e.g. "on_interact", "on_bump", "on_swap"). No-ops if the function
// is not defined in the script. Used by listeners and action handlers.
func CallScriptEvent(eventName string, entity *ecs.Entity, l *world.Level) error {
	if !entity.HasComponent(component.Script) {
		return nil
	}
	if err := ensureInit(entity, l); err != nil {
		return err
	}
	sc := entity.GetComponent(component.Script).(*component.ScriptComponent)
	sc.Ctx.(*scriptContext).Level = l
	return callIfExists(sc, eventName)
}

func callIfExists(sc *component.ScriptComponent, fn string) error {
	if !sc.Interpreter.HasFunction(fn) {
		return nil
	}
	if _, err := sc.Interpreter.Call(fn); err != nil {
		return fmt.Errorf("ScriptSystem: %s: %w", fn, err)
	}
	return nil
}

// scriptContext holds mutable state shared between the system and the closures
// registered on the interpreter. Update Level each tick before calling scripts.
type scriptContext struct {
	Entity *ecs.Entity
	Level  *world.Level
}

func ensureInit(entity *ecs.Entity, l *world.Level) error {
	sc := entity.GetComponent(component.Script).(*component.ScriptComponent)
	if sc.Interpreter != nil {
		return nil
	}
	if err := initScript(sc, entity, l); err != nil {
		return fmt.Errorf("ScriptSystem: init %s: %w", sc.ScriptPath, err)
	}
	return nil
}

func initScript(sc *component.ScriptComponent, entity *ecs.Entity, l *world.Level) error {
	code, err := os.ReadFile(sc.ScriptPath)
	if err != nil {
		return err
	}

	ctx := &scriptContext{Entity: entity, Level: l}
	sc.Ctx = ctx

	interp := basic.NewMechanicalBasic()
	registerScriptFuncs(interp, sc, ctx)

	if err := interp.Load(string(code)); err != nil {
		return err
	}

	sc.Interpreter = interp
	return nil
}

func registerScriptFuncs(interp *basic.MechBasic, sc *component.ScriptComponent, ctx *scriptContext) {
	interp.RegisterFunc("get_x", func(args ...any) (any, error) {
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		return float64(pc.GetX()), nil
	})
	interp.RegisterFunc("get_y", func(args ...any) (any, error) {
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		return float64(pc.GetY()), nil
	})
	interp.RegisterFunc("get_z", func(args ...any) (any, error) {
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		return float64(pc.GetZ()), nil
	})
	interp.RegisterFunc("get_var", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("get_var: expected 1 argument")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("get_var: argument must be a string")
		}
		if sc.Vars == nil {
			return nil, nil
		}
		return sc.Vars[key], nil
	})
	interp.RegisterFunc("set_var", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("set_var: expected 2 arguments")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("set_var: first argument must be a string")
		}
		if sc.Vars == nil {
			sc.Vars = make(map[string]any)
		}
		sc.Vars[key] = args[1]
		return nil, nil
	})
	interp.RegisterFunc("get_flag", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("get_flag: expected 1 argument")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("get_flag: argument must be a string")
		}
		return ctx.Level.Flags[key], nil
	})
	interp.RegisterFunc("set_flag", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("set_flag: expected 2 arguments")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("set_flag: first argument must be a string")
		}
		ctx.Level.Flags[key] = args[1]
		return nil, nil
	})
	interp.RegisterFunc("add_dead", func(args ...any) (any, error) {
		ctx.Entity.AddComponent(&rlcomponents.DeadComponent{})
		return nil, nil
	})
	interp.RegisterFunc("kill_player", func(args ...any) (any, error) {
		for _, e := range ctx.Level.Entities {
			if e != nil && e.HasComponent(component.Player) {
				e.AddComponent(&rlcomponents.DeadComponent{})
				break
			}
		}
		return nil, nil
	})
	interp.RegisterFunc("spawn_entity", func(args ...any) (any, error) {
		if len(args) != 4 {
			return nil, errors.New("spawn_entity: expected 4 arguments (blueprint, x, y, z)")
		}
		blueprint, ok := args[0].(string)
		if !ok {
			return nil, errors.New("spawn_entity: first argument must be a string")
		}
		x := toInt(args[1])
		y := toInt(args[2])
		z := toInt(args[3])
		e, err := factory.Create(blueprint, x, y)
		if err != nil {
			return nil, err
		}
		e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent).SetPosition(x, y, z)
		ctx.Level.AddEntity(e)
		return nil, nil
	})
	interp.RegisterFunc("add_message", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("add_message: expected 1 argument")
		}
		msg, ok := args[0].(string)
		if !ok {
			return nil, errors.New("add_message: argument must be a string")
		}
		message.AddMessage(msg)
		return nil, nil
	})
	interp.RegisterFunc("num_to_str", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("num_to_str: expected 1 argument")
		}
		return fmt.Sprintf("%v", toInt(args[0])), nil
	})
	interp.RegisterFunc("rnd_int", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("rnd_int: expected 1 argument")
		}
		n := toInt(args[0])
		if n <= 0 {
			return float64(0), nil
		}
		return float64(rand.Intn(n)), nil
	})
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
