package factory

import (
	"errors"
	"math/rand/v2"
	"strings"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

// JSON factory for modern blueprints
var jsonFactory = ecs.NewJSONFactory()

func init() {
	registerComponents()
}

// FactoryLoad loads blueprints using the JSON factory only.
func FactoryLoad(folderName string) error {
	return jsonFactory.LoadBlueprintsFromDir(folderName)
}

func Create(name string, x int, y int) (*ecs.Entity, error) {
	// Prefer JSON factory blueprints if available
	if jsonFactory.BlueprintExists(name) {
		entity, err := jsonFactory.CreateWithCallback(name, func(comp ecs.Component) error {
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

		// Preload starting inventory if present (legacy InventoryComponent).
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

		// Preload starting inventory for body-aware inventory.
		if entity.HasComponent(component.BodyInventory) && entity.HasComponent(component.Body) {
			inv := entity.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
			bc := entity.GetComponent(component.Body).(*rlcomponents.BodyComponent)
			if inv.StartingInventory != nil {
				for _, item := range inv.StartingInventory {
					itemEntity, err := Create(item, 0, 0)
					if err != nil {
						return nil, err
					}
					inv.AddItem(itemEntity)
				}
				inv.EquipAllBest(bc)
			}
		}

		skill.Initialize(entity)
		skill.SyncEquippedSkills(entity)
		class.SyncSkills(entity)

		if entity.HasComponent(component.LayeredAppearance) {
			lac := entity.GetComponent(component.LayeredAppearance).(*component.LayeredAppearanceComponent)
			if lac.Randomize {
				bodyTypes := []string{"mid", "slim"}
				lac.BodyType = bodyTypes[rand.IntN(len(bodyTypes))]
				lac.BodyIndex = rand.IntN(5)
				lac.HairIndex = rand.IntN(6) - 1 // -1 = no hair, 0-4 = style
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
