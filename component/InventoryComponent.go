package component

import (
	"fmt"

	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/message"
)

// InventoryComponent .
type InventoryComponent struct {
	LeftHand  *ecs.Entity
	RightHand *ecs.Entity
	Head      *ecs.Entity
	Torso     *ecs.Entity
	Legs      *ecs.Entity
	Feet      *ecs.Entity
	Bag       []*ecs.Entity
}

func (pc InventoryComponent) GetType() string {
	return "InventoryComponent"
}

func (iC *InventoryComponent) AddItem(item *ecs.Entity) {
	iC.Bag = append(iC.Bag, item)
}

func (iC *InventoryComponent) RemoveItem(item *ecs.Entity) {
	for i, v := range iC.Bag {
		if v == item {
			iC.Bag = append(iC.Bag[:i], iC.Bag[i+1:]...)
		}
	}
}

// TODO this is out of scope for a component design wise
func (ic *InventoryComponent) Use(entity *ecs.Entity, target *ecs.Entity) {
	item := entity.GetComponent("ItemComponent").(*ItemComponent)
	if item.Effect == "heal" {
		// TODO MEH...
		target.GetComponent("HealthComponent").(*HealthComponent).Health += item.Value
		ic.RemoveItem(entity)
		message.AddMessage(fmt.Sprint("You healed yourself for ", item.Value))
	} else if item.Slot != "bag" {
		ic.Equip(entity) // Todo - temp equip
	}
}

func (ic *InventoryComponent) Equip(item *ecs.Entity) {
	if item.HasComponent("ItemComponent") {
		itemComponent := item.GetComponent("ItemComponent").(*ItemComponent)

		switch itemComponent.Slot {
		case "hand":
			if ic.RightHand != nil {
				if ic.LeftHand == nil {
					message.AddMessage("Equipped to left hand")
					ic.LeftHand = item
					ic.RemoveItem(item)
				} else {
					message.AddMessage("Equipped to right hand 2")
					ic.AddItem(ic.RightHand)
					ic.RightHand = item
					ic.RemoveItem(item)
				}
			} else {
				message.AddMessage("Equipped to right hand")
				ic.RightHand = item
				ic.RemoveItem(item)
			}
		case "head":
			if ic.Head != nil {
				ic.AddItem(ic.Head)
			}
			ic.Head = item
			ic.RemoveItem(item)
		case "torso":
			if ic.Torso != nil {
				ic.AddItem(ic.Torso)
			}
			ic.Torso = item
			ic.RemoveItem(item)
		case "legs":
			if ic.Legs != nil {
				ic.AddItem(ic.Legs)
			}
			ic.Legs = item
			ic.RemoveItem(item)
		case "feet":
			if ic.Feet != nil {
				ic.AddItem(ic.Feet)
			}
			ic.Feet = item
			ic.RemoveItem(item)
		default:
			message.AddMessage("Not equipable")
		}
	}
}

func (ic *InventoryComponent) GetAttackModifier() int {
	mod := 0
	if ic.RightHand != nil {
		if ic.RightHand.HasComponent("WeaponComponent") {

			item := ic.RightHand.GetComponent("WeaponComponent").(*WeaponComponent)
			mod += item.AttackBonus
		}
	}

	if ic.LeftHand != nil {
		if ic.LeftHand.HasComponent("WeaponComponent") {

			item := ic.LeftHand.GetComponent("WeaponComponent").(*WeaponComponent)
			mod += item.AttackBonus
		}
	}

	return mod
}

func (ic *InventoryComponent) GetAttackDice() string {
	dice := ""
	if ic.RightHand != nil {
		if ic.RightHand.HasComponent("WeaponComponent") {

			item := ic.RightHand.GetComponent("WeaponComponent").(*WeaponComponent)
			if item.AttackDice != "" {
				dice = item.AttackDice
			}
		}
	}

	if ic.LeftHand != nil {
		if ic.LeftHand.HasComponent("WeaponComponent") {

			item := ic.LeftHand.GetComponent("WeaponComponent").(*WeaponComponent)
			if item.AttackDice != "" {
				if dice != "" {
					dice += "+"
				}
				dice += item.AttackDice
			}
		}
	}

	return dice
}

func (ic *InventoryComponent) GetDefenseModifier() int {
	mod := 0
	if ic.Head != nil {
		if ic.Head.HasComponent("ArmorComponent") {
			item := ic.Head.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	if ic.LeftHand != nil {
		if ic.LeftHand.HasComponent("ArmorComponent") {
			item := ic.LeftHand.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	if ic.RightHand != nil {
		if ic.RightHand.HasComponent("ArmorComponent") {
			item := ic.RightHand.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	if ic.Torso != nil {
		if ic.Head.HasComponent("ArmorComponent") {
			item := ic.Torso.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	if ic.Legs != nil {
		if ic.Head.HasComponent("ArmorComponent") {

			item := ic.Legs.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	if ic.Feet != nil {
		if ic.Head.HasComponent("ArmorComponent") {

			item := ic.Feet.GetComponent("ArmorComponent").(*ArmorComponent)
			mod += item.DefenseBonus
		}
	}

	return mod
}
