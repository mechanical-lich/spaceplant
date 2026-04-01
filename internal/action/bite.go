package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/rlbodycombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const bitePoisonDC = 13
const bitePoisonDuration = 5

// BiteAction attacks the solid entity directly in front of the actor for 1d8
// damage. On a successful hit the target must pass a CON saving throw (DC 13)
// or become poisoned for 5 turns. No additional damage is dealt by the save.
type BiteAction struct{}

func (a BiteAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostAttack
}

func (a BiteAction) Available(entity *ecs.Entity, level *world.Level) bool {
	if !entity.HasComponent(component.Direction) {
		return false
	}
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())
	return target != nil && target != entity
}

func (a BiteAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())

	if target != nil && target != entity {
		sc := entity.GetComponent(component.Stats).(*component.StatsComponent)
		orig := sc.BasicAttackDice
		sc.BasicAttackDice = "1d8"
		landed := entityhelpers.Hit(level, entity, target)
		sc.BasicAttackDice = orig

		if landed && !target.HasComponent(component.Dead) && !target.HasComponent(rlcomponents.Poisoned) {
			// CON saving throw — no extra damage, just poison on failure.
			conMod := 0
			if target.HasComponent(rlcomponents.Stats) {
				tsc := target.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
				conMod = rlcombat.GetModifier(tsc.Con)
			}
			tpc := target.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			roll, err := dice.ParseDiceRequest("1d20")
			if err == nil {
				failed := roll.Result+conMod < bitePoisonDC
				if failed {
					target.AddComponent(&rlcomponents.PoisonedComponent{Duration: bitePoisonDuration})
				}
				mlgeevent.GetQueuedInstance().QueueEvent(rlbodycombat.CombatEvent{
					X: tpc.GetX(), Y: tpc.GetY(), Z: tpc.GetZ(),
					Attacker:     entity,
					Defender:     target,
					AttackerName: rlentity.GetName(entity),
					DefenderName: rlentity.GetName(target),
					DamageType:   "poison",
					SaveFail:     failed,
					SavePass:     !failed,
				})
			}
		}
	}

	rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}
