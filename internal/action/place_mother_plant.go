package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// PlaceMotherPlantAction plants the saboteur's mother plant cutting at the
// actor's current tile. It fires a PlaceMotherPlantEventData which the sim
// state listener handles to actually spawn the entity (keeping SimWorld out
// of the action layer). Can only be used once per run.
type PlaceMotherPlantAction struct{}

func (a PlaceMotherPlantAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostMove
}

func (a PlaceMotherPlantAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	if !entity.HasComponent(component.Position) {
		return false
	}
	// The sim listener will enforce the one-use rule via MotherPlantPlaced.
	return true
}

func (a PlaceMotherPlantAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	eventsystem.EventManager.SendEvent(eventsystem.PlaceMotherPlantEventData{
		X: pc.GetX(),
		Y: pc.GetY(),
		Z: pc.GetZ(),
	})
	message.AddMessage("You press the cutting into the deck plating. It begins to pulse.")
	rlenergy.SetActionCost(entity, energy.CostMove)
	_ = level
	return nil
}
