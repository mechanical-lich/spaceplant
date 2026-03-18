package system

import (
	"fmt"

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
				if entity.HasComponent("Inventory") {
					inventory := entity.GetComponent("Inventory").(*component.InventoryComponent)
					used := false
					for _, v := range inventory.Bag {
						item := v.GetComponent("Item").(*component.ItemComponent)
						if item.Effect == "heal" {
							entity.GetComponent("Health").(*component.HealthComponent).Health += item.Value
							inventory.RemoveItem(v)
							used = true

							message.AddMessage(fmt.Sprint("You healed yourself for ", item.Value))
							break
						}
					}
					if !used {
						message.AddMessage("You do not have any healing items")
					}
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
				if entity.HasComponent("Inventory") {
					inventory := entity.GetComponent("Inventory").(*component.InventoryComponent)
					used := false
					for _, v := range inventory.Bag {
						item := v.GetComponent("Item").(*component.ItemComponent)
						if item.Slot != "bag" {
							// TODO MEH...
							inventory.Equip(v)
							used = true

							message.AddMessage(fmt.Sprint("You equipped an item ", v.GetComponent("Description").(*component.DescriptionComponent).Name))
							break
						}
					}
					if !used {
						message.AddMessage("You do not have anything to equip")
					}
				}
			case "P": // Pickup
				if entity.HasComponent("Inventory") {
					inventory := entity.GetComponent("Inventory").(*component.InventoryComponent)
					pc := entity.GetComponent("Position").(*component.PositionComponent)
					var entities []*ecs.Entity
					l.GetEntitiesAt(pc.GetX(), pc.GetY(), z, &entities)
					for _, v := range entities {
						if v.HasComponent("Item") {
							inventory.AddItem(v)
							l.RemoveEntity(v)
							break
						}
					}
				}

			}

			if move(entity, l, deltaX, deltaY) {
				entityHit := l.GetEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY, z)
				if entityHit != nil && entityHit != entity {
					if entityHit != entity {
						hit(l, entity, entityHit)
					}
				}
			}
		}
	}

	return nil
}
