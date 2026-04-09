package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/combat"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// MeleeSpecialAction performs a melee attack against the entity directly in
// front of the actor. On a successful hit (and when ResistDC > 0) the target
// must pass a Cool Check or suffer a status condition as configured via ActionParams.
type MeleeSpecialAction struct {
	Params ActionParams
}

func (a MeleeSpecialAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return a.Params.Cost(energy.CostAttack)
}

func (a MeleeSpecialAction) Available(entity *ecs.Entity, level *world.Level) bool {
	if !entity.HasComponent(component.Direction) {
		return false
	}
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())
	return target != nil && target != entity
}

func (a MeleeSpecialAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())

	if target == nil || target == entity {
		rlenergy.SetActionCost(entity, a.Params.Cost(energy.CostAttack))
		return nil
	}

	verb := a.Params.Verb
	if verb == "" {
		verb = "attack"
	}
	coolDC := a.Params.ResistDC
	statusCondition := a.Params.StatusConditionOnFailSave
	statusDuration := a.Params.StatusConditionDuration
	if statusDuration <= 0 {
		statusDuration = 5
	}

	landed := entityhelpers.Hit(level, entity, target)

	if landed && !target.HasComponent(component.Dead) && coolDC > 0 {
		tpc := target.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		passed := combat.ResistCheck(target, coolDC, a.Params.CheckStat)

		if !passed {
			switch statusCondition {
			case "poison", "burning":
				condDice := a.Params.ExtraDamageOnFailedSave
				if condDice == "" {
					condDice = "1"
				}
				acc := rlcomponents.GetOrCreateActiveConditions(target)
				acc.Add(&rlcomponents.DamageConditionComponent{
					Name:       statusCondition,
					Duration:   statusDuration,
					DamageDice: condDice,
					DamageType: statusCondition,
				})
			case "slowed":
				if !target.HasComponent(rlcomponents.Slowed) {
					target.AddComponent(&rlcomponents.SlowedComponent{Duration: statusDuration})
				}
			case "haste":
				if !target.HasComponent(rlcomponents.Haste) {
					target.AddComponent(&rlcomponents.HasteComponent{Duration: statusDuration})
				}
			}
		}

		attackerName := ""
		if entity.HasComponent(rlcomponents.Description) {
			attackerName = entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
		}
		defenderName := ""
		if target.HasComponent(rlcomponents.Description) {
			defenderName = target.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
		}

		mlgeevent.GetQueuedInstance().QueueEvent(combat.CombatEvent{
			X: tpc.GetX(), Y: tpc.GetY(), Z: tpc.GetZ(),
			Attacker:     entity,
			Defender:     target,
			AttackerName: attackerName,
			DefenderName: defenderName,
			Source:       verb,
			DamageType:   a.Params.DamageType,
			SaveFail:     !passed,
			SavePass:     passed,
		})
	}

	rlenergy.SetActionCost(entity, a.Params.Cost(energy.CostAttack))
	return nil
}
