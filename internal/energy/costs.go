package energy

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// Base action costs. Movement is further modified by terrain.
const (
	CostMove   = 100
	CostAttack = 100
	CostQuick  = 50 // pickup, open door, heal, equip
)

// MoveCost returns the energy cost for an entity to move onto the given tile.
// It multiplies the base movement cost by the tile's MovementCost.
// A tile MovementCost of 0 is treated as 1 (default).
func MoveCost(tile *rlworld.Tile) int {
	def := rlworld.TileDefinitions[tile.Type]
	mult := def.MovementCost
	if mult == 0 {
		mult = 1
	}
	return CostMove * mult
}

// SetActionCost records the energy cost of the action the entity just took.
// CleanUp will read and deduct this value.
func SetActionCost(entity *ecs.Entity, cost int) {
	if entity.HasComponent(component.Energy) {
		ec := entity.GetComponent(component.Energy).(*component.EnergyComponent)
		ec.LastActionCost = cost
	}
}
