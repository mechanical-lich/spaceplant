package system

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/path"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/skill"
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

	if !entity.HasComponent(rlcomponents.MyTurn) {
		if err := ensureInit(entity, l); err != nil {
			log.Printf("[ScriptSystem] failed to init script for %q: %v", entity.Blueprint, err)
		}
		return nil
	}

	// Commit the turn before attempting script init so a broken script doesn't
	// leave the entity holding MyTurn indefinitely and stalling the turn loop.
	entity.AddComponent(rlcomponents.GetTurnTaken())
	if entity.HasComponent(rlcomponents.Energy) {
		ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		rlenergy.SetActionCost(entity, ec.Speed)
	}

	if err := ensureInit(entity, l); err != nil {
		log.Printf("[ScriptSystem] failed to init script for %q: %v", entity.Blueprint, err)
		return err
	}

	sc := entity.GetComponent(component.Script).(*component.ScriptComponent)
	sc.Ctx.(*scriptContext).Level = l
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

	// --- Movement ---

	interp.RegisterFunc("move", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("move: expected 2 arguments (dx, dy)")
		}
		dx, dy := toInt(args[0]), toInt(args[1])
		rlentity.Face(ctx.Entity, dx, dy)
		return nil, action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(ctx.Entity, ctx.Level)
	})

	interp.RegisterFunc("move_random", func(args ...any) (any, error) {
		dx, dy := randomCardinal()
		rlentity.Face(ctx.Entity, dx, dy)
		return nil, action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(ctx.Entity, ctx.Level)
	})

	interp.RegisterFunc("move_toward", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("move_toward: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		dx, dy := scriptPathStep(ctx.Entity, tx, ty, ctx.Level)
		rlentity.Face(ctx.Entity, dx, dy)
		err := action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(ctx.Entity, ctx.Level)
		if dx == 0 && dy == 0 {
			return float64(0), err
		}
		return float64(1), err
	})

	interp.RegisterFunc("move_away_from", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("move_away_from: expected 2 arguments (fx, fy)")
		}
		fx, fy := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		x, y := pc.GetX(), pc.GetY()
		dx, dy := 0, 0
		if x > fx {
			dx = 1
		} else if x < fx {
			dx = -1
		}
		if y > fy {
			dy = 1
		} else if y < fy {
			dy = -1
		}
		if dx == 0 && dy == 0 {
			dx, dy = randomCardinal()
		} else {
			sepX := math.Abs(float64(x - fx))
			sepY := math.Abs(float64(y - fy))
			if sepX > sepY {
				dy = 0
			} else if sepY > sepX {
				dx = 0
			}
		}
		rlentity.Face(ctx.Entity, dx, dy)
		return nil, action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(ctx.Entity, ctx.Level)
	})

	// --- Combat ---

	interp.RegisterFunc("attack", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("attack: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		fdx, fdy := facingDelta(pc.GetX(), pc.GetY(), tx, ty)
		rlentity.Face(ctx.Entity, fdx, fdy)
		return nil, action.AttackAction{TargetX: tx, TargetY: ty}.Execute(ctx.Entity, ctx.Level)
	})

	interp.RegisterFunc("use_skill", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("use_skill: expected 1 argument (ai_type)")
		}
		aiType, ok := args[0].(string)
		if !ok {
			return nil, errors.New("use_skill: argument must be a string")
		}
		_, act := skill.SkillForAIType(ctx.Entity, aiType)
		if act == nil {
			return float64(0), nil
		}
		return float64(1), act.Execute(ctx.Entity, ctx.Level)
	})

	// --- Sensing ---

	// find_nearest_enemy(range) — scans for the nearest non-friendly Body entity
	// within range tiles. Returns the integer distance, or -1 if none found.
	// Caches target position in sc.Vars for get_target_x/y.
	interp.RegisterFunc("find_nearest_enemy", func(args ...any) (any, error) {
		sightRange := 8
		if len(args) >= 1 {
			sightRange = toInt(args[0])
		}
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		var nearby []*ecs.Entity
		ctx.Level.GetEntitiesAround(pc.GetX(), pc.GetY(), pc.GetZ(), sightRange, sightRange, &nearby)

		var closest *ecs.Entity
		closestDist := math.MaxFloat64
		for _, e := range nearby {
			if e == ctx.Entity || rlcombat.IsFriendly(ctx.Entity, e) {
				continue
			}
			if e.HasComponent(rlcomponents.Dead) || !e.HasComponent(rlcomponents.Body) {
				continue
			}
			ep := e.GetComponent(component.Position).(*component.PositionComponent)
			dx := float64(ep.GetX() - pc.GetX())
			dy := float64(ep.GetY() - pc.GetY())
			d := math.Sqrt(dx*dx + dy*dy)
			if d < closestDist {
				closest = e
				closestDist = d
			}
		}

		if closest == nil {
			if sc.Vars == nil {
				sc.Vars = make(map[string]any)
			}
			sc.Vars["_target_x"] = float64(0)
			sc.Vars["_target_y"] = float64(0)
			sc.Vars["_target_dist"] = float64(-1)
			return float64(-1), nil
		}

		ep := closest.GetComponent(component.Position).(*component.PositionComponent)
		if sc.Vars == nil {
			sc.Vars = make(map[string]any)
		}
		sc.Vars["_target_x"] = float64(ep.GetX())
		sc.Vars["_target_y"] = float64(ep.GetY())
		sc.Vars["_target_dist"] = closestDist
		return closestDist, nil
	})

	interp.RegisterFunc("get_target_x", func(args ...any) (any, error) {
		if sc.Vars == nil {
			return float64(0), nil
		}
		if v, ok := sc.Vars["_target_x"]; ok {
			return v, nil
		}
		return float64(0), nil
	})

	interp.RegisterFunc("get_target_y", func(args ...any) (any, error) {
		if sc.Vars == nil {
			return float64(0), nil
		}
		if v, ok := sc.Vars["_target_y"]; ok {
			return v, nil
		}
		return float64(0), nil
	})

	interp.RegisterFunc("distance_to", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("distance_to: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		dx := float64(tx - pc.GetX())
		dy := float64(ty - pc.GetY())
		return math.Sqrt(dx*dx + dy*dy), nil
	})

	interp.RegisterFunc("has_los", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("has_los: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		if rlfov.Los(ctx.Level.Level, pc.GetX(), pc.GetY(), tx, ty, pc.GetZ()) {
			return float64(1), nil
		}
		return float64(0), nil
	})

	// is_aligned(tx, ty) — returns 1 if entity shares a row or column with (tx,ty)
	// and has clear LOS; 0 otherwise.
	interp.RegisterFunc("is_aligned", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("is_aligned: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		ax, ay, z := pc.GetX(), pc.GetY(), pc.GetZ()
		if ax != tx && ay != ty {
			return float64(0), nil
		}
		if !rlfov.Los(ctx.Level.Level, ax, ay, tx, ty, z) {
			return float64(0), nil
		}
		return float64(1), nil
	})

	// friendly_on_ray(tx, ty) — returns 1 if a friendly entity lies on the
	// straight line between self and (tx, ty); 0 otherwise.
	interp.RegisterFunc("friendly_on_ray", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("friendly_on_ray: expected 2 arguments (tx, ty)")
		}
		tx, ty := toInt(args[0]), toInt(args[1])
		pc := ctx.Entity.GetComponent(component.Position).(*component.PositionComponent)
		ax, ay, z := pc.GetX(), pc.GetY(), pc.GetZ()
		dx, dy := 0, 0
		if tx > ax {
			dx = 1
		} else if tx < ax {
			dx = -1
		}
		if ty > ay {
			dy = 1
		} else if ty < ay {
			dy = -1
		}
		steps := toInt(args[0]) // reuse tx as depth when axis-aligned
		if dx != 0 {
			steps = int(math.Abs(float64(tx - ax)))
		} else {
			steps = int(math.Abs(float64(ty - ay)))
		}
		for i := 1; i <= steps; i++ {
			var entities []*ecs.Entity
			ctx.Level.GetEntitiesAt(ax+dx*i, ay+dy*i, z, &entities)
			for _, e := range entities {
				if e != ctx.Entity && rlcombat.IsFriendly(ctx.Entity, e) {
					return float64(1), nil
				}
			}
		}
		return float64(0), nil
	})

	interp.RegisterFunc("get_faction", func(args ...any) (any, error) {
		if ctx.Entity.HasComponent(component.Description) {
			dc := ctx.Entity.GetComponent(component.Description).(*component.DescriptionComponent)
			return dc.Faction, nil
		}
		return "", nil
	})
}

// randomCardinal returns a random non-zero cardinal direction step.
func randomCardinal() (int, int) {
	dx := rand.Intn(3) - 1
	dy := 0
	if dx == 0 {
		dy = rand.Intn(3) - 1
	}
	return dx, dy
}

// facingDelta returns a unit step from (fx, fy) toward (tx, ty).
func facingDelta(fx, fy, tx, ty int) (int, int) {
	dx, dy := 0, 0
	if fx < tx {
		dx = 1
	} else if fx > tx {
		dx = -1
	}
	if fy < ty {
		dy = 1
	} else if fy > ty {
		dy = -1
	}
	return dx, dy
}

// scriptPathStep returns the cardinal (dx, dy) step toward (tx, ty) using A*,
// with SizedGraph support for large entities. Returns (0,0) if already there or stuck.
func scriptPathStep(entity *ecs.Entity, tx, ty int, level *world.Level) (int, int) {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	z := pc.GetZ()
	fromX, fromY := pc.GetX(), pc.GetY()

	var graph path.Graph = level.Level
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		w, h := sc.Width, sc.Height
		if w > 1 || h > 1 {
			graph = &rlworld.SizedGraph{Level: level.Level, Width: w, Height: h, Entity: entity}
			startX := fromX - w/2
			startY := fromY - h/2
			fromX = max(startX, min(tx, startX+w-1))
			fromY = max(startY, min(ty, startY+h-1))
		}
	}

	from := level.Level.GetTilePtr(fromX, fromY, z)
	to := level.Level.GetTilePtr(tx, ty, z)
	if from == nil || to == nil {
		return randomCardinal()
	}

	steps, _, _ := path.Path(graph, from.Idx, to.Idx)
	if len(steps) < 2 {
		return 0, 0
	}

	s0 := level.Level.GetTilePtrIndex(steps[0])
	s1 := level.Level.GetTilePtrIndex(steps[1])
	sx, sy, _ := s0.Coords()
	nx, ny, _ := s1.Coords()
	return nx - sx, ny - sy
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
