package action

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// HealAction uses the first healing item from the entity's inventory.
type HealAction struct{}

func (a HealAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a HealAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	return findHealItem(entity) != nil
}

func (a HealAction) Execute(entity *ecs.Entity, _ *world.Level) error {
	item := findHealItem(entity)
	if item == nil {
		message.AddMessage("You do not have any healing items")
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}
	ic := item.GetComponent(component.Item).(*component.ItemComponent)
	entityhelpers.HealBodyParts(entity, ic.Value)
	removeFromInventory(entity, item)
	message.AddMessage(fmt.Sprint("You used a health pack (", ic.Value, " HP spread across damaged parts)"))
	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}

// findHealItem returns the first heal-effect item in the entity's inventory, or nil.
func findHealItem(entity *ecs.Entity) *ecs.Entity {
	bag, _ := inventoryBag(entity)
	for _, v := range bag {
		if !v.HasComponent(component.Item) {
			continue
		}
		ic := v.GetComponent(component.Item).(*component.ItemComponent)
		if ic.Effect == "heal" {
			return v
		}
	}
	return nil
}

// removeFromInventory removes an item from whichever inventory the entity has.
func removeFromInventory(entity, item *ecs.Entity) {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		inv.RemoveItem(item)
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.RemoveItem(item)
	}
}

// inventoryBag returns the entity's item bag and a bool indicating if it was found.
func inventoryBag(entity *ecs.Entity) ([]*ecs.Entity, bool) {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		return inv.Bag, true
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		return inv.Bag, true
	}
	return nil, false
}
