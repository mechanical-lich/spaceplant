package game

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

func playerBag(player *ecs.Entity) []*ecs.Entity {
	if player.HasComponent(component.BodyInventory) {
		return player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Bag
	}
	if player.HasComponent(component.Inventory) {
		return player.GetComponent(component.Inventory).(*component.InventoryComponent).Bag
	}
	return nil
}

func playerRemoveItem(player *ecs.Entity, item *ecs.Entity) {
	if player.HasComponent(component.BodyInventory) {
		player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).RemoveItem(item)
		return
	}
	if player.HasComponent(component.Inventory) {
		player.GetComponent(component.Inventory).(*component.InventoryComponent).RemoveItem(item)
	}
}

func playerEquipItem(player *ecs.Entity, item *ecs.Entity) {
	if player.HasComponent(component.BodyInventory) && player.HasComponent(component.Body) {
		inv := player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		bc := player.GetComponent(component.Body).(*component.BodyComponent)
		inv.AutoEquip(item, bc)
		return
	}
	if player.HasComponent(component.Inventory) {
		player.GetComponent(component.Inventory).(*component.InventoryComponent).Equip(item)
	}
}

// playerEquipped returns a slot→item map regardless of inventory type.
func playerEquipped(player *ecs.Entity) map[string]*ecs.Entity {
	if player.HasComponent(component.BodyInventory) {
		return player.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent).Equipped
	}
	if player.HasComponent(component.Inventory) {
		inv := player.GetComponent(component.Inventory).(*component.InventoryComponent)
		m := map[string]*ecs.Entity{}
		if inv.Head != nil {
			m["Head"] = inv.Head
		}
		if inv.Torso != nil {
			m["Torso"] = inv.Torso
		}
		if inv.Legs != nil {
			m["Legs"] = inv.Legs
		}
		if inv.Feet != nil {
			m["Feet"] = inv.Feet
		}
		if inv.RightHand != nil {
			m["R Hand"] = inv.RightHand
		}
		if inv.LeftHand != nil {
			m["L Hand"] = inv.LeftHand
		}
		return m
	}
	return nil
}
