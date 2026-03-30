package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ShoveAction pushes the solid entity directly in front of the actor one tile
// further in the same direction, provided the tile behind the target is free.
// Massive and inanimate entities cannot be shoved.
type ShoveAction struct{}

func (a ShoveAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostAttack
}

func (a ShoveAction) Available(entity *ecs.Entity, level *world.Level) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ())
	return target != nil && target != entity &&
		!target.HasComponent(component.Massive) &&
		!target.HasComponent(component.Inanimate)
}

func (a ShoveAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	dx, dy := directionDeltas(dc.Direction)
	z := pc.GetZ()

	target := level.GetSolidEntityAt(pc.GetX()+dx, pc.GetY()+dy, z)
	if target != nil && target != entity &&
		!target.HasComponent(component.Massive) &&
		!target.HasComponent(component.Inanimate) {

		targetPC := target.GetComponent(component.Position).(*component.PositionComponent)
		destX := targetPC.GetX() + dx
		destY := targetPC.GetY() + dy

		destTile := level.Level.GetTilePtr(destX, destY, z)
		if destTile != nil && !destTile.IsSolid() &&
			level.GetSolidEntityAt(destX, destY, z) == nil {
			level.Level.PlaceEntity(destX, destY, z, target)
		}
	}

	rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}

// directionDeltas converts a DirectionComponent direction value to a (dx, dy) step.
// Direction encoding mirrors MoveAction.Execute: 0=East, 1=South, 2=North, 3=West.
func directionDeltas(dir int) (int, int) {
	switch dir {
	case 0:
		return 1, 0
	case 1:
		return 0, 1
	case 2:
		return 0, -1
	default: // 3
		return -1, 0
	}
}
