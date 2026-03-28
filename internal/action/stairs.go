package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// StairsAction uses stairs at the entity's current position.
type StairsAction struct{}

func (a StairsAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a StairsAction) Available(entity *ecs.Entity, level *world.Level) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	tile := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), pc.GetZ())
	if tile == nil {
		return false
	}
	def := rlworld.TileDefinitions[tile.Type]
	return def.StairsDown || def.StairsUp
}

func (a StairsAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	tile := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), pc.GetZ())
	if tile != nil {
		def := rlworld.TileDefinitions[tile.Type]
		if def.StairsDown {
			eventsystem.EventManager.SendEvent(eventsystem.StairsEventData{})
		}
		if def.StairsUp {
			eventsystem.EventManager.SendEvent(eventsystem.StairsEventData{Up: true})
		}
	}
	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}
