package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// DecayingComponent is implemented by status effects that expire over time.
type DecayingComponent interface {
	Decay() bool
	GetType() ecs.ComponentType
}

type StatusConditionSystem struct {
}

var statusConditions = []ecs.ComponentType{"Poisoned", "Alerted", "Haste", "Slowed"}

func (s StatusConditionSystem) UpdateSystem(data any) error {
	return nil
}

func (s StatusConditionSystem) Requires() []ecs.ComponentType {
	return nil
}

// StatusConditionSystem ticks every decaying status on entities whose turn it is.
// SpeedModifier conditions are applied on first tick and reverted on expiry.
func (s StatusConditionSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	if entity.HasComponent("MyTurn") {
		for _, statusCondition := range statusConditions {
			if entity.HasComponent(statusCondition) {
				dc := entity.GetComponent(statusCondition).(DecayingComponent)

				// Apply speed-modifying effects once.
				if sm, ok := dc.(rlcomponents.SpeedModifier); ok {
					sm.ApplyOnce(entity)
				}

				if dc.Decay() {
					// Revert speed-modifying effects before removing.
					if sm, ok := dc.(rlcomponents.SpeedModifier); ok {
						sm.Revert(entity)
					}
					entity.RemoveComponent(statusCondition)
				}
			}
		}
	}
	return nil
}
