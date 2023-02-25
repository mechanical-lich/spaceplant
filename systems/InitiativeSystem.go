package systems

import (
	"github.com/mechanical-lich/game-engine/entity"
	"github.com/mechanical-lich/spaceplant/components"
)

type InitiativeSystem struct {
}

// InitiativeSystem .
func (s InitiativeSystem) Update(levelInterface interface{}, entity *entity.Entity) error {
	if entity.HasComponent("InitiativeComponent") {
		ic := entity.GetComponent("InitiativeComponent").(*components.InitiativeComponent)
		ic.Ticks--

		if ic.Ticks <= 0 {
			ic.Ticks = ic.DefaultValue
			if ic.OverrideValue > 0 {
				ic.Ticks = ic.OverrideValue
			}

			if !entity.HasComponent("MyTurnComponent") {

				mTC := &components.MyTurnComponent{}
				entity.AddComponent(mTC)

			}
		}
	}

	return nil
}
