package systems

import (
	"github.com/mechanical-lich/spaceplant/components"
	"github.com/mechanical-lich/spaceplant/level"
)

type CleanUpSystem struct {
}

// CleanUpSystem .
func (s CleanUpSystem) Update(level *level.Level) {
	for _, entity := range level.Entities {
		if entity.HasComponent("MyTurnComponent") {
			entity.RemoveComponent("MyTurnComponent")
		}

		if entity.HasComponent("DeadComponent") {
			entity.RemoveComponent("AttackComponent")
			if entity.HasComponent("FoodComponent") {
				fc := entity.GetComponent("FoodComponent").(*components.FoodComponent)
				if fc.Amount <= 0 {
					level.RemoveEntity(entity)
				}
			} else {
				level.RemoveEntity(entity)
			}
		}

	}

}
