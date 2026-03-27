package initiative

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// AdvanceInitiative decrements the initiative tick counter for every entity
// that has an InitiativeComponent and grants MyTurn to those whose counter
// reaches zero. Returns (playerGotTurn, anyGotTurn).
func AdvanceInitiative(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool) {
	for _, entity := range entities {
		if !entity.HasComponent("Initiative") {
			continue
		}
		ic := entity.GetComponent("Initiative").(*component.InitiativeComponent)
		ic.Ticks--

		if ic.Ticks <= 0 {
			ic.Ticks = ic.DefaultValue
			if ic.OverrideValue > 0 {
				ic.Ticks = ic.OverrideValue
			}

			entity.AddComponent(rlcomponents.GetMyTurn())
			entity.RemoveComponent(rlcomponents.TurnTaken)

			anyGotTurn = true
			if entity == player {
				playerGotTurn = true
			}
		}
	}
	return
}
