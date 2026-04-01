package system

import (
	"math"
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/path"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type AISystem struct {
	Watcher *ecs.Entity
}

func (s *AISystem) UpdateSystem(data any) error {
	return nil
}

func (s *AISystem) Requires() []ecs.ComponentType {
	return nil
}

// UpdateEntity runs AI decision logic for one NPC tick, then executes the
// chosen action using the same shared action types the player uses.
func (s *AISystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}
	if !entity.HasComponent(rlcomponents.MyTurn) {
		return nil
	}

	isAI := entity.HasComponent(component.WanderAI) ||
		entity.HasComponent(component.HostileAI) ||
		entity.HasComponent(component.DefensiveAI)
	if !isAI {
		return nil
	}

	level := levelInterface.(*world.Level)
	entity.AddComponent(rlcomponents.GetTurnTaken())

	if rlentity.HandleDeath(entity) {
		rlentity.CheckDeathAnnouncement(s.Watcher, entity, level.Level)
		return nil
	}

	pc := entity.GetComponent(component.Position).(*component.PositionComponent)

	// --- Defensive AI: strike back at whoever attacked us. ---
	if entity.HasComponent(component.DefensiveAI) {
		aic := entity.GetComponent(component.DefensiveAI).(*component.DefensiveAIComponent)
		if aic.Attacked {
			act := action.AttackAction{TargetX: aic.AttackerX, TargetY: aic.AttackerY}
			dx, dy := facingDelta(pc.GetX(), pc.GetY(), aic.AttackerX, aic.AttackerY)
			rlentity.Face(entity, dx, dy)
			return act.Execute(entity, level)
		}
	}

	// --- Hostile AI: pathfind to the nearest food entity. ---
	if entity.HasComponent(component.HostileAI) {
		hc := entity.GetComponent(component.HostileAI).(*component.HostileAIComponent)

		// Always resolve movement first so TargetX/Y stays current each tick.
		// This also gives us the fallback (dx, dy) if we don't shoot.
		dx, dy := hostileDirection(entity, hc, pc, level)

		// If the entity has a skill with ai_type "align_and_shoot", try to fire.
		// Only shoots when aligned, in range, and with clear LOS.
		if def, act := skill.SkillForAIType(entity, "align_and_shoot"); def != nil {
			if tryAlignAndShoot(entity, pc, hc, def, act, level) {
				return nil
			}
		}

		// If the entity has a melee skill, use it instead of a plain move when adjacent.
		if _, act := skill.SkillForAIType(entity, "melee_skill"); act != nil {
			tdx := hc.TargetX - pc.GetX()
			tdy := hc.TargetY - pc.GetY()
			if tdx < 0 {
				tdx = -tdx
			}
			if tdy < 0 {
				tdy = -tdy
			}
			if tdx <= 1 && tdy <= 1 && (tdx != 0 || tdy != 0) {
				fdx, fdy := facingDelta(pc.GetX(), pc.GetY(), hc.TargetX, hc.TargetY)
				rlentity.Face(entity, fdx, fdy)
				return act.Execute(entity, level)
			}
		}

		rlentity.Face(entity, dx, dy)
		err := action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)
		// Hostile NPCs also eat any food entities at the target tile.
		eatFoodAt(entity, level, pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())
		return err
	}

	// --- Wander AI: random cardinal step. ---
	if entity.HasComponent(component.WanderAI) {
		dx, dy := randomCardinal()
		rlentity.Face(entity, dx, dy)
		return action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)
	}

	return nil
}

// hostileDirection returns the (dx, dy) step toward the nearest food entity,
// falling back to a random wander step when nothing is visible.
func hostileDirection(entity *ecs.Entity, hc *component.HostileAIComponent,
	pc *component.PositionComponent, level *world.Level) (int, int) {

	z := pc.GetZ()
	var nearby []*ecs.Entity
	level.GetEntitiesAround(pc.GetX(), pc.GetY(), z, hc.SightRange, hc.SightRange, &nearby)

	closest := (*ecs.Entity)(nil)
	distance := 999999.0
	for _, e := range nearby {
		if e == entity || rlcombat.IsFriendly(entity, e) {
			continue
		}
		if !e.HasComponent(component.Food) || e.HasComponent(rlcomponents.Dead) {
			continue
		}
		foodPC := e.GetComponent(component.Position).(*component.PositionComponent)
		from := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
		to := level.Level.GetTilePtr(foodPC.GetX(), foodPC.GetY(), z)
		if from == nil || to == nil {
			continue
		}
		if d := level.Level.PathEstimate(from.Idx, to.Idx); d < distance {
			closest = e
			distance = d
		}
	}

	if closest == nil {
		return randomCardinal()
	}

	foodPC := closest.GetComponent(component.Position).(*component.PositionComponent)
	hc.TargetX = foodPC.GetX()
	hc.TargetY = foodPC.GetY()

	fromX, fromY := pc.GetX(), pc.GetY()
	var graph path.Graph = level.Level
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		w, h := sc.Width, sc.Height
		if w > 1 || h > 1 {
			graph = &rlworld.SizedGraph{Level: level.Level, Width: w, Height: h, Entity: entity}
			startX := fromX - w/2
			startY := fromY - h/2
			fromX = max(startX, min(hc.TargetX, startX+w-1))
			fromY = max(startY, min(hc.TargetY, startY+h-1))
		}
	}

	from := level.Level.GetTilePtr(fromX, fromY, z)
	to := level.Level.GetTilePtr(hc.TargetX, hc.TargetY, z)
	if from == nil || to == nil {
		return randomCardinal()
	}

	steps, _, _ := path.Path(graph, from.Idx, to.Idx)
	hc.Path = append(hc.Path[:0], steps...)
	if len(steps) < 2 {
		return randomCardinal()
	}

	s0 := level.Level.GetTilePtrIndex(steps[0])
	s1 := level.Level.GetTilePtrIndex(steps[1])
	sx, sy, _ := s0.Coords()
	nx, ny, _ := s1.Coords()
	return nx - sx, ny - sy
}

// eatFoodAt calls rlentity.Eat for each food entity at the given tile.
func eatFoodAt(entity *ecs.Entity, level *world.Level, x, y, z int) {
	var entities []*ecs.Entity
	level.GetEntitiesAt(x, y, z, &entities)
	for _, e := range entities {
		if e.HasComponent(component.Food) {
			rlentity.Eat(entity, e)
		}
	}
}

// randomCardinal returns a random non-zero cardinal direction step.
func randomCardinal() (int, int) {
	dx := rand.Intn(3) - 1 // -1, 0, 1
	dy := 0
	if dx == 0 {
		dy = rand.Intn(3) - 1
	}
	return dx, dy
}

// tryAlignAndShoot checks whether the entity is lined up with its target and
// within the skill's range. If so it faces the target and fires the action,
// returning true. Returns false so the caller can fall through to movement.
func tryAlignAndShoot(entity *ecs.Entity, pc *component.PositionComponent,
	hc *component.HostileAIComponent, def *skill.SkillDef, act action.Action,
	level *world.Level) bool {

	ax, ay, z := pc.GetX(), pc.GetY(), pc.GetZ()
	tx, ty := hc.TargetX, hc.TargetY

	depth := def.ActionParams.Depth
	if depth <= 0 {
		depth = 3
	}

	dx, dy := 0, 0
	dist := 0

	switch {
	case ax == tx && int(math.Abs(float64(ty-ay))) <= depth:
		// Same column, target within range.
		dist = int(math.Abs(float64(ty - ay)))
		if ty < ay {
			dy = -1
		} else {
			dy = 1
		}
	case ay == ty && int(math.Abs(float64(tx-ax))) <= depth:
		// Same row, target within range.
		dist = int(math.Abs(float64(tx - ax)))
		if tx < ax {
			dx = -1
		} else {
			dx = 1
		}
	default:
		return false
	}

	if dist == 0 {
		return false
	}

	// Confirm line of sight is clear.
	if !rlfov.Los(level.Level, ax, ay, tx, ty, z) {
		return false
	}

	rlentity.Face(entity, dx, dy)
	act.Execute(entity, level) //nolint:errcheck
	return true
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
