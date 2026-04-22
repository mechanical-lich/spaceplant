package action

import (

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// RoundhouseKickAction attacks kicks all solid entities in the 8 tiles surrounding the entity.
type RoundhouseKickAction struct {
}

func (a RoundhouseKickAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostAttack * 2
}

func (a RoundhouseKickAction) Available(entity *ecs.Entity, level *world.Level) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)

	targetOffsets := [][2]int{
		{0, -1},  // North
		{1, -1},  // Northeast
		{1, 0},   // Southeast
		{1, 1},   // South
		{0, 1},   // Southwest
		{-1, 1},  // West
		{-1, 0},  // Northwest
		{-1, -1}, // North
	}
	for _, offset := range targetOffsets {
		tx := pc.GetX() + offset[0]
		ty := pc.GetY() + offset[1]
		target := level.GetSolidEntityAt(tx, ty, pc.GetZ())
		if target != nil && target != entity {
			return true
		}
	}
	return false
}

func (a RoundhouseKickAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	targetOffsets := [][2]int{
		{0, -1},  // North
		{1, -1},  // Northeast
		{1, 0},   // Southeast
		{1, 1},   // South
		{0, 1},   // Southwest
		{-1, 1},  // West
		{-1, 0},  // Northwest
		{-1, -1}, // North
	}
	for _, offset := range targetOffsets {
		tx := pc.GetX() + offset[0]
		ty := pc.GetY() + offset[1]
		target := level.GetSolidEntityAt(tx, ty, pc.GetZ())
		if target != nil && target != entity {
			entityhelpers.Hit(level, entity, target)
		}

		level.AddTileAnim(tx, ty, pc.GetZ(), &world.TileAnim{
			Resource: "fx", SpriteX: 0, SpriteY: 0,
			FrameCount: 4, FrameSpeed: 4, TTL: 10,
		})
	}

rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}
