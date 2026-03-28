package initiative

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// AdvanceEnergy adds Speed to Energy for every entity with an EnergyComponent.
// When Energy reaches the Threshold the entity receives MyTurn.
// Returns (playerGotTurn, anyGotTurn).
func AdvanceEnergy(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool) {
	for _, entity := range entities {
		if !entity.HasComponent(component.Energy) {
			continue
		}
		ec := entity.GetComponent(component.Energy).(*component.EnergyComponent)
		ec.Energy += ec.Speed

		if ec.Energy >= ec.Threshold && !entity.HasComponent(rlcomponents.MyTurn) {
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
