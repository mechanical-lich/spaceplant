package system

import (
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type CleanUpSystem struct {
}

// CleanUpSystem .
func (s CleanUpSystem) Update(level *world.Level) {
	for _, entity := range level.Entities {
		if entity.HasComponent("MyTurn") {
			entity.RemoveComponent("MyTurn")
		}

		if entity.HasComponent("Dead") {
			entity.RemoveComponent("AttackComponent")
			entity.RemoveComponent("Solid")
		}
	}
}
