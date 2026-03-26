package system

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type PlayerSystem struct {
}

func (s *PlayerSystem) UpdateSystem(data any) error {
	return nil
}

func (s *PlayerSystem) Requires() []ecs.ComponentType {
	return nil
}

// PlayerSystem .
func (s *PlayerSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	l := levelInterface.(*world.Level)

	if entity.HasComponent("PlayerComponent") {
		if entity.HasComponent("MyTurn") {
			pc := entity.GetComponent("Position").(*component.PositionComponent)
			dc := entity.GetComponent("Direction").(*component.DirectionComponent)
			playerComponent := entity.GetComponent("PlayerComponent").(*component.PlayerComponent)
			command := playerComponent.PopCommand()

			if command != "" {
				entity.AddComponent(rlcomponents.GetTurnTaken())
			}

			z := pc.GetZ()
			deltaX := 0
			deltaY := 0

			switch command {
			case "W":
				deltaY--
				dc.Direction = 2
			case "S":
				deltaY++
				dc.Direction = 1
			case "A":
				deltaX--
				dc.Direction = 3
			case "D":
				deltaX++
				dc.Direction = 0
			case "F":
				direction := playerComponent.PopCommand()
				if direction == "" {
					message.AddMessage("Wasn't given a direction to shoot!")
				} else {
					message.AddMessage("Shoot in the " + direction + " direction!")
				}
			case "H":
				used := false
				var bag []*ecs.Entity
				var removeFn func(*ecs.Entity) bool
				if entity.HasComponent(component.BodyInventory) {
					inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
					bag = inv.Bag
					removeFn = inv.RemoveItem
				} else if entity.HasComponent(component.Inventory) {
					inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
					bag = inv.Bag
					removeFn = inv.RemoveItem
				}
				for _, v := range bag {
					if !v.HasComponent(component.Item) {
						continue
					}
					item := v.GetComponent(component.Item).(*component.ItemComponent)
					if item.Effect == "heal" {
						healBodyParts(entity, item.Value)
						removeFn(v)
						used = true
						message.AddMessage(fmt.Sprint("You used a health pack (", item.Value, " HP spread across damaged parts)"))
						break
					}
				}
				if !used {
					message.AddMessage("You do not have any healing items")
				}
			case "Period": // Stairs
				tile := l.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
				if tile != nil {
					def := rlworld.TileDefinitions[tile.Type]
					if def.StairsDown {
						eventsystem.EventManager.SendEvent(eventsystem.StairsEventData{})
					}
					if def.StairsUp {
						eventsystem.EventManager.SendEvent(eventsystem.StairsEventData{Up: true})
					}
				}

			case "E":
				used := false
				if entity.HasComponent(component.BodyInventory) && entity.HasComponent(component.Body) {
					inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
					bc := entity.GetComponent(component.Body).(*component.BodyComponent)
					for _, v := range inv.Bag {
						if !v.HasComponent(component.Item) {
							continue
						}
						item := v.GetComponent(component.Item).(*component.ItemComponent)
						if item.Slot != component.BagSlot {
							inv.AutoEquip(v, bc)
							used = true
							message.AddMessage(fmt.Sprint("You equipped ", v.GetComponent(component.Description).(*component.DescriptionComponent).Name))
							break
						}
					}
				} else if entity.HasComponent(component.Inventory) {
					inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
					for _, v := range inv.Bag {
						if !v.HasComponent(component.Item) {
							continue
						}
						item := v.GetComponent(component.Item).(*component.ItemComponent)
						if item.Slot != component.BagSlot {
							inv.Equip(v)
							used = true
							message.AddMessage(fmt.Sprint("You equipped ", v.GetComponent(component.Description).(*component.DescriptionComponent).Name))
							break
						}
					}
				}
				if !used {
					message.AddMessage("You do not have anything to equip")
				}
			case "P": // Pickup
				var entities []*ecs.Entity
				l.GetEntitiesAt(pc.GetX(), pc.GetY(), z, &entities)
				if entity.HasComponent(component.BodyInventory) {
					inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
					for _, v := range entities {
						if v.HasComponent(component.Item) {
							inv.AddItem(v)
							l.RemoveEntity(v)
							break
						}
					}
				} else if entity.HasComponent(component.Inventory) {
					inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
					for _, v := range entities {
						if v.HasComponent(component.Item) {
							inv.AddItem(v)
							l.RemoveEntity(v)
							break
						}
					}
				}

			}

			if move(entity, l, deltaX, deltaY) {
				// Prefer non-door entities (e.g. monsters) over doors so that
				// bumping into a monster standing in a doorway attacks rather
				// than toggling the door.
				var candidates []*ecs.Entity
				l.GetEntitiesAt(pc.GetX()+deltaX, pc.GetY()+deltaY, z, &candidates)
				var entityHit, doorHit *ecs.Entity
				for _, e := range candidates {
					if e == entity {
						continue
					}
					if e.HasComponent(component.Door) {
						if doorHit == nil {
							doorHit = e
						}
					} else if entityHit == nil {
						entityHit = e
					}
				}
				if entityHit == nil {
					entityHit = doorHit
				}

				if entityHit != nil {
					if rlentity.CheckInteraction(entity, entityHit) {
						// interaction consumed the bump — do not attack or swap
					} else if entityHit.HasComponent(component.Door) {
						toggleDoor(entity, entityHit)
					} else if rlcombat.IsFriendly(entity, entityHit) {
						rlentity.CheckExcuseMe(entityHit)
						hit(l, entity, entityHit)
					} else {
						hit(l, entity, entityHit)
					}
				}
			} else if deltaX != 0 || deltaY != 0 {
				rlentity.CheckPassOver(entity, l.Level, pc.GetX(), pc.GetY(), z)
			}
		}
	}

	return nil
}

// toggleDoor opens or closes a door entity, with locked/faction checks.
// It also immediately syncs the appearance sprite so the change is visible
// this tick (door entities are iterated before the player, so DoorSystem
// would otherwise lag one tick behind).
func toggleDoor(actor, doorEntity *ecs.Entity) {
	door := doorEntity.GetComponent(component.Door).(*component.DoorComponent)
	if door.Open {
		door.Open = false
	} else if door.Locked {
		message.AddMessage("The door is locked.")
		return
	} else if door.OwnedBy != "" {
		if !actor.HasComponent(component.Description) {
			message.AddMessage("Access denied.")
			return
		}
		desc := actor.GetComponent(component.Description).(*component.DescriptionComponent)
		if desc.Faction != door.OwnedBy {
			message.AddMessage("Access denied.")
			return
		}
		door.Open = true
	} else {
		door.Open = true
	}
	rlsystems.SyncDoorAppearance(doorEntity, component.Appearance)
}
