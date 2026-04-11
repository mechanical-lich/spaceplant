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

// Action name constants used for last-action repeat penalty.
const (
	advActWander          = "wander"
	advActMove            = "move"
	advActMelee           = "melee"
	advActMeleeSkill      = "melee_skill"
	advActShoot           = "shoot"
	advActFlee            = "flee"
	advActSpreadOvergrowth = "spread_overgrowth"
)

// AdvancedAISystem drives entities with an AdvancedAIComponent.
// It supports idle / hunt / flee states, proximity-scaled aggression/fear
// rolls, randomness-based action scoring, and optional friendly-fire avoidance.
type AdvancedAISystem struct {
	Watcher *ecs.Entity
}

func (s *AdvancedAISystem) UpdateSystem(data any) error {
	return nil
}

func (s *AdvancedAISystem) Requires() []ecs.ComponentType {
	return nil
}

func (s *AdvancedAISystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}
	if !entity.HasComponent(rlcomponents.MyTurn) {
		return nil
	}
	if !entity.HasComponent(component.AdvancedAI) {
		return nil
	}

	level := levelInterface.(*world.Level)
	entity.AddComponent(rlcomponents.GetTurnTaken())

	if rlentity.HandleDeath(entity) {
		rlentity.CheckDeathAnnouncement(s.Watcher, entity, level.Level)
		return nil
	}

	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	aic := entity.GetComponent(component.AdvancedAI).(*component.AdvancedAIComponent)

	// Find nearest food target within sight range.
	target, dist := advFindTarget(entity, aic, pc, level)

	// Run the state machine transition logic.
	advTransitionState(aic, target, dist)

	switch aic.State {
	case "flee":
		dx, dy := advFleeDirection(pc, aic)
		rlentity.Face(entity, dx, dy)
		aic.LastAction = advActFlee
		return action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)

	case "hunt":
		if target == nil {
			// Target lost; wander until something comes back into view.
			dx, dy := randomCardinal()
			rlentity.Face(entity, dx, dy)
			aic.LastAction = advActWander
			return action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)
		}
		targetPC := target.GetComponent(component.Position).(*component.PositionComponent)
		aic.TargetX, aic.TargetY = targetPC.GetX(), targetPC.GetY()
		return advExecuteHunt(entity, aic, pc, dist, level)

	default: // idle
		// If the entity has a spread_overgrowth skill, alternate between spreading
		// and wandering. The repeat penalty naturally encourages alternation.
		if _, act := skill.SkillForAIType(entity, advActSpreadOvergrowth); act != nil {
			spreadScore := 55
			wanderScore := 45
			if aic.LastAction == advActSpreadOvergrowth {
				spreadScore -= 30
			} else if aic.LastAction == advActWander {
				wanderScore -= 30
			}
			if spreadScore >= wanderScore {
				aic.LastAction = advActSpreadOvergrowth
				return act.Execute(entity, level)
			}
		}
		dx, dy := randomCardinal()
		rlentity.Face(entity, dx, dy)
		aic.LastAction = advActWander
		return action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)
	}
}

// advTransitionState updates aic.State each tick based on target visibility,
// proximity, aggressiveness, and fear rolls.
func advTransitionState(aic *component.AdvancedAIComponent, target *ecs.Entity, dist int) {
	if aic.State == "" {
		aic.State = "idle"
	}

	// Proximity bonus increases the effective aggressiveness when the target is
	// close.  At max sight range the bonus is 0; at point-blank it peaks at ~40.
	proximityBonus := 0
	if target != nil && aic.SightRange > 0 {
		bonus := (aic.SightRange - dist) * 2
		if bonus < 0 {
			bonus = 0
		}
		if bonus > 40 {
			bonus = 40
		}
		proximityBonus = bonus
	}

	effectiveAgg := aic.Aggressiveness + proximityBonus
	effectiveFear := aic.Fear

	switch aic.State {
	case "idle":
		if target == nil {
			return
		}
		// Whichever disposition is dominant takes priority.
		if effectiveAgg >= effectiveFear {
			if rand.Intn(100) < effectiveAgg {
				aic.State = "hunt"
			}
		} else {
			if rand.Intn(100) < effectiveFear {
				advSetFleeTarget(aic, target)
				aic.State = "flee"
			}
		}

	case "hunt":
		if target == nil {
			aic.State = "idle"
			return
		}
		// Fear can override aggressiveness when it dominates.
		if effectiveFear > effectiveAgg && rand.Intn(100) < effectiveFear {
			advSetFleeTarget(aic, target)
			aic.State = "flee"
		}

	case "flee":
		if target != nil {
			// Keep the flee-from position current while the threat is visible.
			advSetFleeTarget(aic, target)
		}
		// Calm down once the threat is gone or far enough away.
		if (target == nil || dist > aic.SightRange) && rand.Intn(100) < 30 {
			aic.State = "idle"
		}
	}
}

// advSetFleeTarget records the target's current position as the point to flee from.
func advSetFleeTarget(aic *component.AdvancedAIComponent, target *ecs.Entity) {
	if !target.HasComponent(component.Position) {
		return
	}
	tpc := target.GetComponent(component.Position).(*component.PositionComponent)
	aic.FleeFromX, aic.FleeFromY = tpc.GetX(), tpc.GetY()
}

// advExecuteHunt scores available actions and executes the highest-scoring one.
// Randomness adds jitter to each score; the last-used action receives a penalty
// to encourage variety.
func advExecuteHunt(entity *ecs.Entity, aic *component.AdvancedAIComponent,
	pc *component.PositionComponent, dist int, level *world.Level) error {

	jitter := func() int {
		if aic.Randomness <= 0 {
			return 0
		}
		return rand.Intn(aic.Randomness + 1)
	}
	penalty := func(name string) int {
		if aic.LastAction == name {
			return 30
		}
		return 0
	}

	type candidate struct {
		name  string
		score int
	}

	// Wander is always available as a fallback.
	best := candidate{advActWander, 10 + jitter() - penalty(advActWander)}

	// Path toward target.
	dx, dy := advHuntDirection(entity, aic, pc, level)
	if s := 60 + jitter() - penalty(advActMove); s > best.score {
		best = candidate{advActMove, s}
	}

	// Melee attack when adjacent.
	if dist <= 1 {
		if s := 80 + jitter() - penalty(advActMelee); s > best.score {
			best = candidate{advActMelee, s}
		}
	}

	// Melee skill (e.g. poisonous bite): prefer over plain melee when adjacent.
	var meleeSkillAct action.Action
	if _, act := skill.SkillForAIType(entity, "melee_skill"); act != nil {
		if dist <= 1 {
			if s := 85 + jitter() - penalty(advActMeleeSkill); s > best.score {
				best = candidate{advActMeleeSkill, s}
				meleeSkillAct = act
			}
		}
	}

	// Ranged skill: align and shoot.
	var shootDX, shootDY int
	var shootAct action.Action
	if def, act := skill.SkillForAIType(entity, "align_and_shoot"); def != nil {
		sdx, sdy, aligned := advAligned(pc, aic, def)
		if aligned {
			depth := def.ActionParams.Depth
			if depth <= 0 {
				depth = 3
			}
			losOK := rlfov.Los(level.Level, pc.GetX(), pc.GetY(), aic.TargetX, aic.TargetY, pc.GetZ())
			noFF := !aic.AvoidsFriendlyFire || !advFriendlyOnRay(entity, pc, sdx, sdy, depth, level)
			if losOK && noFF {
				if s := 70 + jitter() - penalty(advActShoot); s > best.score {
					best = candidate{advActShoot, s}
					shootDX, shootDY = sdx, sdy
					shootAct = act
				}
			}
		}
	}

	// Execute the winning action.
	switch best.name {
	case advActShoot:
		rlentity.Face(entity, shootDX, shootDY)
		shootAct.Execute(entity, level) //nolint:errcheck
		aic.LastAction = advActShoot
		return nil

	case advActMeleeSkill:
		fdx, fdy := facingDelta(pc.GetX(), pc.GetY(), aic.TargetX, aic.TargetY)
		rlentity.Face(entity, fdx, fdy)
		aic.LastAction = advActMeleeSkill
		meleeSkillAct.Execute(entity, level) //nolint:errcheck
		return nil

	case advActMelee:
		fdx, fdy := facingDelta(pc.GetX(), pc.GetY(), aic.TargetX, aic.TargetY)
		rlentity.Face(entity, fdx, fdy)
		aic.LastAction = advActMelee
		return action.AttackAction{TargetX: aic.TargetX, TargetY: aic.TargetY}.Execute(entity, level)

	case advActMove:
		rlentity.Face(entity, dx, dy)
		err := action.MoveAction{DeltaX: dx, DeltaY: dy}.Execute(entity, level)
		eatFoodAt(entity, level, pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())
		aic.LastAction = advActMove
		return err

	default: // wander
		wdx, wdy := randomCardinal()
		rlentity.Face(entity, wdx, wdy)
		aic.LastAction = advActWander
		return action.MoveAction{DeltaX: wdx, DeltaY: wdy}.Execute(entity, level)
	}
}

// advFindTarget returns the nearest food entity within sight range and its
// integer Euclidean distance, or (nil, 0) when none is visible.
func advFindTarget(entity *ecs.Entity, aic *component.AdvancedAIComponent,
	pc *component.PositionComponent, level *world.Level) (*ecs.Entity, int) {

	z := pc.GetZ()
	var nearby []*ecs.Entity
	level.GetEntitiesAround(pc.GetX(), pc.GetY(), z, aic.SightRange, aic.SightRange, &nearby)

	var closest *ecs.Entity
	closestDist := math.MaxFloat64

	for _, e := range nearby {
		if e == entity || rlcombat.IsFriendly(entity, e) {
			continue
		}
		if e.HasComponent(rlcomponents.Dead) || !e.HasComponent(rlcomponents.Body) {
			continue
		}
		ep := e.GetComponent(component.Position).(*component.PositionComponent)
		d := advEuclidean(pc.GetX(), pc.GetY(), ep.GetX(), ep.GetY())
		if d < closestDist {
			closest = e
			closestDist = d
		}
	}

	if closest == nil {
		return nil, 0
	}
	return closest, int(closestDist)
}

// advHuntDirection computes the next step toward aic.TargetX/Y using
// A* pathfinding, with SizedGraph support for large entities.
func advHuntDirection(entity *ecs.Entity, aic *component.AdvancedAIComponent,
	pc *component.PositionComponent, level *world.Level) (int, int) {

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
			fromX = max(startX, min(aic.TargetX, startX+w-1))
			fromY = max(startY, min(aic.TargetY, startY+h-1))
		}
	}

	from := level.Level.GetTilePtr(fromX, fromY, z)
	to := level.Level.GetTilePtr(aic.TargetX, aic.TargetY, z)
	if from == nil || to == nil {
		return randomCardinal()
	}

	steps, _, _ := path.Path(graph, from.Idx, to.Idx)
	aic.Path = append(aic.Path[:0], steps...)
	if len(steps) < 2 {
		return randomCardinal()
	}

	s0 := level.Level.GetTilePtrIndex(steps[0])
	s1 := level.Level.GetTilePtrIndex(steps[1])
	sx, sy, _ := s0.Coords()
	nx, ny, _ := s1.Coords()
	return nx - sx, ny - sy
}

// advFleeDirection returns the cardinal step that moves most directly away
// from aic.FleeFromX/Y.
func advFleeDirection(pc *component.PositionComponent, aic *component.AdvancedAIComponent) (int, int) {
	x, y := pc.GetX(), pc.GetY()
	fx, fy := aic.FleeFromX, aic.FleeFromY

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
		return randomCardinal()
	}

	// Prefer movement along the axis with the greatest separation.
	sepX := math.Abs(float64(x - fx))
	sepY := math.Abs(float64(y - fy))
	if sepX > sepY {
		dy = 0
	} else if sepY > sepX {
		dx = 0
	}

	return dx, dy
}

// advAligned checks whether the entity is in the same row or column as its
// target and within the skill's depth range.  Returns the facing delta and
// whether the entity is aligned.
func advAligned(pc *component.PositionComponent, aic *component.AdvancedAIComponent,
	def *skill.SkillDef) (int, int, bool) {

	ax, ay := pc.GetX(), pc.GetY()
	tx, ty := aic.TargetX, aic.TargetY

	depth := def.ActionParams.Depth
	if depth <= 0 {
		depth = 3
	}

	switch {
	case ax == tx && int(math.Abs(float64(ty-ay))) <= depth:
		dy := 1
		if ty < ay {
			dy = -1
		}
		return 0, dy, true
	case ay == ty && int(math.Abs(float64(tx-ax))) <= depth:
		dx := 1
		if tx < ax {
			dx = -1
		}
		return dx, 0, true
	default:
		return 0, 0, false
	}
}

// advFriendlyOnRay returns true if any friendly entity occupies a tile on the
// straight line from the entity in direction (dx, dy) for up to depth steps.
func advFriendlyOnRay(entity *ecs.Entity, pc *component.PositionComponent,
	dx, dy, depth int, level *world.Level) bool {

	x, y, z := pc.GetX(), pc.GetY(), pc.GetZ()
	for i := 1; i <= depth; i++ {
		var entities []*ecs.Entity
		level.GetEntitiesAt(x+dx*i, y+dy*i, z, &entities)
		for _, e := range entities {
			if e != entity && rlcombat.IsFriendly(entity, e) {
				return true
			}
		}
	}
	return false
}

// advEuclidean returns the Euclidean distance between two integer points.
func advEuclidean(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}
