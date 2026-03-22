package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type CleanUpSystem struct {
}

// Update strips transient components from entities each frame.
func (s CleanUpSystem) Update(level *world.Level) {
	for _, entity := range level.Entities {
		if entity.HasComponent(rlcomponents.MyTurn) && entity.HasComponent(rlcomponents.TurnTaken) {
			entity.RemoveComponent(rlcomponents.MyTurn)
			entity.RemoveComponent(rlcomponents.TurnTaken)
		}

		if entity.HasComponent(rlcomponents.Dead) {
			entity.RemoveComponent(component.Attack)
			entity.RemoveComponent(rlcomponents.Solid)
		}
	}
}
