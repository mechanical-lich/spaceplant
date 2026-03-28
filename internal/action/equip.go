package action

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// EquipAction auto-equips the first equippable item in the entity's inventory.
type EquipAction struct{}

func (a EquipAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a EquipAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	return findEquippableItem(entity) != nil
}

func (a EquipAction) Execute(entity *ecs.Entity, _ *world.Level) error {
	item := findEquippableItem(entity)
	if item == nil {
		message.AddMessage("You do not have anything to equip")
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	if entity.HasComponent(component.BodyInventory) && entity.HasComponent(component.Body) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		bc := entity.GetComponent(component.Body).(*component.BodyComponent)
		inv.AutoEquip(item, bc)
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.Equip(item)
	}

	name := item.GetComponent(component.Description).(*component.DescriptionComponent).Name
	message.AddMessage(fmt.Sprint("You equipped ", name))
	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}

// findEquippableItem returns the first non-bag item in the entity's inventory.
func findEquippableItem(entity *ecs.Entity) *ecs.Entity {
	bag, _ := inventoryBag(entity)
	for _, v := range bag {
		if !v.HasComponent(component.Item) {
			continue
		}
		ic := v.GetComponent(component.Item).(*component.ItemComponent)
		if ic.Slot != component.BagSlot {
			return v
		}
	}
	return nil
}
