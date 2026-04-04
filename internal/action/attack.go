package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// handsOnlyPen calculates the Penetration value for an unarmed hands_only attack.
// Scales with Physique: PH 10 → 5 Pen, PH 18 → 7 Pen.
func handsOnlyPen(entity *ecs.Entity) int {
	ph := 10
	if entity.HasComponent(component.Stats) {
		ph = entity.GetComponent(component.Stats).(*component.StatsComponent).PH
	}
	return 3 + ph/4
}

// AttackAction attacks the solid entity at (TargetX, TargetY) on the entity's Z level.
type AttackAction struct {
	TargetX, TargetY int
}

func (a AttackAction) Cost(entity *ecs.Entity, _ *world.Level) int {
	if entityHasSkill(entity, "hands_only") && !hasWeaponEquipped(entity) {
		return energy.CostQuick
	}
	return energy.CostAttack
}

func (a AttackAction) Available(entity *ecs.Entity, level *world.Level) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	target := level.GetSolidEntityAt(a.TargetX, a.TargetY, pc.GetZ())
	return target != nil && target != entity
}

func (a AttackAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	target := level.GetSolidEntityAt(a.TargetX, a.TargetY, pc.GetZ())
	cost := energy.CostAttack
	if target != nil && target != entity {
		if entityHasSkill(entity, "hands_only") && !hasWeaponEquipped(entity) {
			entityhelpers.HitWithPen(level, entity, target, handsOnlyPen(entity))
			cost = energy.CostQuick
		} else {
			entityhelpers.Hit(level, entity, target)
		}
	}
	rlenergy.SetActionCost(entity, cost)
	return nil
}

// entityHasSkill checks the entity's SkillComponent directly, avoiding an
// import cycle between the action and skill packages.
func entityHasSkill(entity *ecs.Entity, skillID string) bool {
	if !entity.HasComponent(component.Skill) {
		return false
	}
	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	for _, s := range sc.Skills {
		if s == skillID {
			return true
		}
	}
	return false
}

// hasWeaponEquipped returns true when the entity has a weapon item equipped in
// any hand / body slot.
func hasWeaponEquipped(entity *ecs.Entity) bool {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				return true
			}
		}
		return false
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		return (inv.RightHand != nil && inv.RightHand.HasComponent(component.Weapon)) ||
			(inv.LeftHand != nil && inv.LeftHand.HasComponent(component.Weapon))
	}
	return false
}
