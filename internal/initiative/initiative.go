package initiative

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// PlayerTookTurn advances initiative for all entities after the player acts.
func PlayerTookTurn(entities []*ecs.Entity) {
	for _, entity := range entities {
		if entity.HasComponent("Initiative") {
			ic := entity.GetComponent("Initiative").(*component.InitiativeComponent)
			ic.Ticks--

			if ic.Ticks <= 0 {
				ic.Ticks = ic.DefaultValue
				if ic.OverrideValue > 0 {
					ic.Ticks = ic.OverrideValue
				}

				entity.AddComponent(rlcomponents.GetMyTurn())
				entity.RemoveComponent(rlcomponents.TurnTaken)
			}
		}
	}
}
