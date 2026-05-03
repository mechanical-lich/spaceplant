package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlaction"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// Action is a turn-based action parameterised to spaceplant's Level.
type Action = rlaction.Action[world.Level]

// HasAvailableAction returns true if at least one action in the list is both
// available and affordable given the entity's current energy.
func HasAvailableAction(entity *ecs.Entity, level *world.Level, actions []Action) bool {
	return rlaction.HasAvailableAction(entity, level, actions)
}
