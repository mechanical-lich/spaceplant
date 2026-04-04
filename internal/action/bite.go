package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/rlbodycombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// MeleeSpecialAction performs a melee attack against the entity directly in
// front of the actor. On a successful hit (and when SaveDC > 0) the target
// must pass a saving throw or suffer extra damage and/or a status condition as
// configured via ActionParams.
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

	damageDice := a.Params.DamageDice
	if damageDice == "" {
		damageDice = "1d8"
	}
	damageType := a.Params.DamageType
	saveDC := a.Params.SaveDC
	saveStat := a.Params.SaveStat
	if saveStat == "" {
		saveStat = "con"
	}
	verb := a.Params.Verb
	if verb == "" {
		verb = "attack"
	}
	extraDamage := a.Params.ExtraDamageOnFailedSave
	statusCondition := a.Params.StatusConditionOnFailSave
	statusDuration := a.Params.StatusConditionDuration
	if statusDuration <= 0 {
		statusDuration = 5
	}

	sc := entity.GetComponent(component.Stats).(*component.StatsComponent)
	orig := sc.BasicAttackDice
	sc.BasicAttackDice = damageDice
	landed := entityhelpers.Hit(level, entity, target)
	sc.BasicAttackDice = orig

	if landed && !target.HasComponent(component.Dead) && saveDC > 0 {
		saveStatMod := 0
		if target.HasComponent(rlcomponents.Stats) {
			tsc := target.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
			switch saveStat {
			case "str":
				saveStatMod = rlcombat.GetModifier(tsc.Str)
			case "dex":
				saveStatMod = rlcombat.GetModifier(tsc.Dex)
			case "con":
				saveStatMod = rlcombat.GetModifier(tsc.Con)
			case "int":
				saveStatMod = rlcombat.GetModifier(tsc.Int)
			case "wis":
				saveStatMod = rlcombat.GetModifier(tsc.Wis)
			default:
				saveStatMod = rlcombat.GetModifier(tsc.Con)
			}
		}

		tpc := target.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		roll, err := dice.ParseDiceRequest("1d20")
		if err == nil {
			failed := roll.Result+saveStatMod < saveDC

			if failed {
				if extraDamage != "" {
					dmg, rollErr := dice.Roll(extraDamage)
					if rollErr == nil && dmg > 0 {
						if target.HasComponent(rlcomponents.Health) {
							hc := target.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
							hc.Health -= dmg
							if hc.Health <= 0 {
								hc.Health = 0
								target.AddComponent(&rlcomponents.DeadComponent{})
							}
						}
					}
				}

				switch statusCondition {
				case "poison", "burning":
					condDice := extraDamage
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
				// TODO - Add more status conditions that modify stats
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

			mlgeevent.GetQueuedInstance().QueueEvent(rlbodycombat.CombatEvent{
				X: tpc.GetX(), Y: tpc.GetY(), Z: tpc.GetZ(),
				Attacker:     entity,
				Defender:     target,
				AttackerName: rlentity.GetName(entity),
				DefenderName: rlentity.GetName(target),
				Source:       verb,
				DamageType:   damageType,
				SaveFail:     failed,
				SavePass:     !failed,
			})
		}
	}

	rlenergy.SetActionCost(entity, a.Params.Cost(energy.CostAttack))
	return nil
}
