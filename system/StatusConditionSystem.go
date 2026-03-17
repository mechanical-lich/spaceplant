package system

import (
	"github.com/mechanical-lich/mlge/ecs"
)

// DecayingComponent is implemented by status effects that expire over time.
type DecayingComponent interface {
	Decay() bool
	GetType() ecs.ComponentType
}

type StatusConditionSystem struct {
}

var statusConditions = []ecs.ComponentType{"Poisoned", "Alerted"}

func (s StatusConditionSystem) UpdateSystem(data any) error {
	return nil
}

func (s StatusConditionSystem) Requires() []ecs.ComponentType {
	return nil
}

// StatusConditionSystem .
func (s StatusConditionSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	for _, statusCondition := range statusConditions {
		if entity.HasComponent(statusCondition) {
			pc := entity.GetComponent(statusCondition).(DecayingComponent)

			if pc.Decay() {
				entity.RemoveComponent(statusCondition)
			}
		}
	}
	return nil
}
