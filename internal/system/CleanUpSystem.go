package system

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type CleanUpSystem struct {
}

// Update strips transient components from entities each frame.
func (s CleanUpSystem) Update(level *world.Level) {
	var toRemove []*ecs.Entity

	for _, entity := range level.Entities {
		rlenergy.ResolveTurn(entity)

		if entity.HasComponent(rlcomponents.Dead) {
			if entity.Blueprint == "mother_plant" && entity.HasComponent(rlcomponents.Solid) {
				// Solid still present = first cleanup frame; fire win exactly once.
				eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{
					Outcome: "extermination",
					Message: "The mother plant is dead. The infestation collapses.",
				})
			}
			if entity.Blueprint == "mobile_mother_plant" && entity.HasComponent(rlcomponents.Solid) {
				// Remove the cutting on death — it leaves no corpse.
				toRemove = append(toRemove, entity)
			}
			entity.RemoveComponent(component.Attack)
			entity.RemoveComponent(rlcomponents.Solid)
			entity.RemoveComponent(rlcomponents.MyTurn)
			entity.RemoveComponent(rlcomponents.TurnTaken)
			entity.RemoveComponent(rlcomponents.Energy)

			if entity.HasComponent(rlcomponents.Drops) {
				fmt.Println("Processing drops for", entity.Blueprint)
				drops := entity.GetComponent(rlcomponents.Drops).(*rlcomponents.DropsComponent)
				droppedItems := drops.GetDrops()
				fmt.Println("Dropped items:", droppedItems)
				pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
				for itemName, quantity := range droppedItems {
					for i := 0; i < quantity; i++ {
						item, err := factory.Create(itemName, pc.X, pc.Y)
						if err == nil {
							fmt.Println("Dropping", itemName, "at", pc.X, pc.Y)
							level.AddEntity(item)
						}
					}
				}

				entity.RemoveComponent(rlcomponents.Drops)
			}
		}
	}

	for _, entity := range toRemove {
		level.Level.RemoveEntity(entity)
	}
}
