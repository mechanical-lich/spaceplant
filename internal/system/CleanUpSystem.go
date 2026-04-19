package system

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type CleanUpSystem struct {
	// CtxProvider, if set, is called to build the eval context for win-condition
	// checks on entity death. If nil, no win conditions are evaluated from kills.
	CtxProvider func() wincondition.EvalContext
}

// Update strips transient components from entities each frame.
func (s CleanUpSystem) Update(level *world.Level) {
	var toRemove []*ecs.Entity

	for _, entity := range level.Entities {
		rlenergy.ResolveTurn(entity)

		if entity.HasComponent(rlcomponents.Dead) {
			if entity.HasComponent(rlcomponents.Solid) && s.CtxProvider != nil {
				// Solid still present = first cleanup frame; evaluate kill rules once.
				ctx := s.CtxProvider()
				if rule, ok := wincondition.Active().EvalKill(entity.Blueprint, ctx); ok {
					wincondition.FireRule(rule, "")
				}
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
