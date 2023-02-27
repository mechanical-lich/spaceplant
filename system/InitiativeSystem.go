package system

import (
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
)

type InitiativeSystem struct {
}

// InitiativeSystem .
func (s InitiativeSystem) Update(levelInterface interface{}, entity *ecs.Entity) error {
	if entity.HasComponent("InitiativeComponent") {
		ic := entity.GetComponent("InitiativeComponent").(*component.InitiativeComponent)
		ic.Ticks--

		if ic.Ticks <= 0 {
			ic.Ticks = ic.DefaultValue
			if ic.OverrideValue > 0 {
				ic.Ticks = ic.OverrideValue
			}

			if !entity.HasComponent("MyTurnComponent") {

				mTC := &component.MyTurnComponent{}
				entity.AddComponent(mTC)

			}
		}
	}

	return nil
}
