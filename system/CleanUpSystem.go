package system

import (
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
			entity.RemoveComponent("SolidComponent")

			// if entity.HasComponent("FoodComponent") {
			// 	fc := entity.GetComponent("FoodComponent").(*component.FoodComponent)
			// 	if fc.Amount <= 0 {
			// 		level.DeleteEntity(entity)
			// 	}
			// } else {
			// 	level.DeleteEntity(entity)
			// }
		}

	}

}
