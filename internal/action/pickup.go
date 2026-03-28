package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// PickupAction picks up the first item at the entity's current position.
type PickupAction struct{}

func (a PickupAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a PickupAction) Available(entity *ecs.Entity, level *world.Level) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	var entities []*ecs.Entity
	level.GetEntitiesAt(pc.GetX(), pc.GetY(), pc.GetZ(), &entities)
	for _, v := range entities {
		if v != entity && v.HasComponent(component.Item) {
			return true
		}
	}
	return false
}

func (a PickupAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	z := pc.GetZ()
	var entities []*ecs.Entity
	level.GetEntitiesAt(pc.GetX(), pc.GetY(), z, &entities)

	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, v := range entities {
			if v.HasComponent(component.Item) {
				inv.AddItem(v)
				level.RemoveEntity(v)
				break
			}
		}
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		for _, v := range entities {
			if v.HasComponent(component.Item) {
				inv.AddItem(v)
				level.RemoveEntity(v)
				break
			}
		}
	}

	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}
