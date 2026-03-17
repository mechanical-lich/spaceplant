package factory

import (
	"errors"
	"strings"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/component"
)

// JSON factory for modern blueprints
var jsonFactory = ecs.NewJSONFactory()

func init() {
	registerComponents()
}

// FactoryLoad loads blueprints using the JSON factory only.
func FactoryLoad(filename string) error {
	return jsonFactory.LoadBlueprintsFromFile(filename)
}

func Create(name string, x int, y int) (*ecs.Entity, error) {
	// Prefer JSON factory blueprints if available
	if jsonFactory.BlueprintExists(name) {
		entity, err := jsonFactory.CreateWithCallback(name, func(comp ecs.Component) error {
			// Auto-initialize health if present
			if hc, ok := comp.(*rlcomponents.HealthComponent); ok {
				if hc.Health == 0 {
					hc.Health = hc.MaxHealth
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		// Add position and direction
		pc := &component.PositionComponent{}
		pc.SetPosition(x, y, 0)
		entity.AddComponent(pc)
		entity.AddComponent(&component.DirectionComponent{Direction: 0})

		// Preload starting inventory if present
		if entity.HasComponent(component.Inventory) {
			inv := entity.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
			if inv.StartingInventory != nil {
				for _, item := range inv.StartingInventory {
					itemEntity, err := Create(item, 0, 0)
					if err != nil {
						return nil, err
					}
					inv.AddItem(itemEntity)
				}
				inv.EquipAllBest()
			}
		}

		return entity, nil
	}
	return nil, errors.New("no blueprint found")
}

func ImportInventory(importString string, target *ecs.Entity) error {
	if !target.HasComponent(component.Inventory) {
		return errors.New("no inventory")
	}

	inventory := target.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
	imports := strings.Split(importString, ",")
	for _, v := range imports {
		entity, err := Create(v, 0, 0)
		if err != nil {
			return err
		}
		inventory.AddItem(entity)
		inventory.Equip(entity)
	}

	return nil
}
