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

// PickupItemAction picks up a specific item from a specific tile (which may be
// adjacent to the player, not just underfoot).
type PickupItemAction struct {
	Item          *ecs.Entity
	TileX, TileY, TileZ int
}

func (a PickupItemAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a PickupItemAction) Available(_ *ecs.Entity, _ *world.Level) bool {
	return a.Item != nil
}

func (a PickupItemAction) Execute(entity *ecs.Entity, level *world.Level) error {
	if a.Item == nil {
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	name := ""
	if a.Item.HasComponent(component.Item) {
		name = a.Item.GetComponent(component.Item).(*component.ItemComponent).Name
	}

	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		inv.AddItem(a.Item)
		level.RemoveEntity(a.Item)
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.AddItem(a.Item)
		level.RemoveEntity(a.Item)
	}

	if name != "" {
		message.AddMessage(fmt.Sprintf("You pick up the %s.", name))
	}

	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}

// EquipItemAction picks up a specific item from a specific tile and immediately
// equips it, replacing anything in the same slot.
type EquipItemAction struct {
	Item          *ecs.Entity
	TileX, TileY, TileZ int
}

func (a EquipItemAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a EquipItemAction) Available(_ *ecs.Entity, _ *world.Level) bool {
	return a.Item != nil
}

func (a EquipItemAction) Execute(entity *ecs.Entity, level *world.Level) error {
	if a.Item == nil {
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	name := ""
	if a.Item.HasComponent(component.Item) {
		name = a.Item.GetComponent(component.Item).(*component.ItemComponent).Name
	}

	// Pick up first.
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		inv.AddItem(a.Item)
		level.RemoveEntity(a.Item)
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.AddItem(a.Item)
		level.RemoveEntity(a.Item)
	}

	// Then equip.
	if entity.HasComponent(component.BodyInventory) && entity.HasComponent(component.Body) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		bc := entity.GetComponent(component.Body).(*component.BodyComponent)
		inv.AutoEquip(a.Item, bc)
	} else if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		inv.Equip(a.Item)
	}

	if name != "" {
		message.AddMessage(fmt.Sprintf("You pick up and equip the %s.", name))
	}

	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}
