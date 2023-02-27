package system

import (
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
)

type StatusConditionSystem struct {
}

var statusConditions = []string{"Poisoned", "Alerted"}

// StatusConditionSystem .
func (s StatusConditionSystem) Update(levelInterface interface{}, entity *ecs.Entity) error {
	for _, statusCondition := range statusConditions {
		if entity.HasComponent(statusCondition + "Component") {
			pc := entity.GetComponent(statusCondition + "Component").(component.DecayingComponent)

			if pc.Decay() {
				entity.RemoveComponent(statusCondition + "Component")
			}
		}
	}
	return nil
}
