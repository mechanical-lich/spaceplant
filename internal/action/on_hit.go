package action

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/combat"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// applyOnHitCondition runs the weapon's resistance check against the target and,
// on failure, applies the configured status condition.
func applyOnHitCondition(target *ecs.Entity, wc *component.WeaponComponent) {
	dc := wc.OnHitResistDC
	if dc <= 0 {
		return
	}
	duration := wc.OnHitDuration
	if duration <= 0 {
		duration = 3
	}

	passed := combat.ResistCheck(target, dc, wc.OnHitCheckStat)
	if passed {
		return
	}

	name := ""
	if target.HasComponent(component.Description) {
		name = target.GetComponent(component.Description).(*component.DescriptionComponent).Name
	}

	switch wc.OnHitCondition {
	case "slowed":
		if !target.HasComponent(rlcomponents.Slowed) {
			target.AddComponent(&rlcomponents.SlowedComponent{Duration: duration})
		}
		if name != "" {
			message.AddMessage(fmt.Sprintf("%s is slowed!", name))
		}
	case "poison", "burning":
		acc := rlcomponents.GetOrCreateActiveConditions(target)
		acc.Add(&rlcomponents.DamageConditionComponent{
			Name:       wc.OnHitCondition,
			Duration:   duration,
			DamageDice: "1",
			DamageType: wc.OnHitCondition,
		})
		if name != "" {
			message.AddMessage(fmt.Sprintf("%s is %s!", name, wc.OnHitCondition+"ed"))
		}
	}
}
