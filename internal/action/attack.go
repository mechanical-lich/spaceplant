package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// AttackAction attacks the solid entity at (TargetX, TargetY) on the entity's Z level.
type AttackAction struct {
	TargetX, TargetY int
}

func (a AttackAction) Cost(_ *ecs.Entity, _ *world.Level) int {
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
	if target != nil && target != entity {
		entityhelpers.Hit(level, entity, target)
	}
	rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}
