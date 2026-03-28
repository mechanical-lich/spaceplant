package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// Action represents a single entity action with its associated energy cost.
// Execute applies the action and calls rlenergy.SetActionCost internally.
type Action interface {
	// Cost returns the expected energy cost, used to determine affordability.
	Cost(entity *ecs.Entity, level *world.Level) int
	// Available returns true if this action can be performed right now.
	Available(entity *ecs.Entity, level *world.Level) bool
	// Execute performs the action. It must call rlenergy.SetActionCost.
	Execute(entity *ecs.Entity, level *world.Level) error
}

// HasAvailableAction returns true if at least one action in the list is both
// available and affordable given the entity's current energy. This is used to
// decide whether to wait for player input or skip the entity's turn.
func HasAvailableAction(entity *ecs.Entity, level *world.Level, actions []Action) bool {
	if !entity.HasComponent(rlcomponents.Energy) {
		return false
	}
	ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	for _, a := range actions {
		if a.Available(entity, level) && ec.Energy >= a.Cost(entity, level) {
			return true
		}
	}
	return false
}
