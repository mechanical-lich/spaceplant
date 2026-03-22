package generation

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// spawnDoor replaces a tile door placement with a door entity.
// The underlying tile is set to floor so the entity provides all blocking/visual.
func spawnDoor(l *world.Level, x, y, z int) {
	l.SetTileTypeAt(x, y, z, world.TypeFloor)

	e := &ecs.Entity{}
	pc := &rlcomponents.PositionComponent{}
	pc.SetPosition(x, y, z)
	e.AddComponent(pc)
	e.AddComponent(&rlcomponents.SolidComponent{})
	e.AddComponent(&rlcomponents.DoorComponent{
		ClosedSpriteX: 11 * config.SpriteWidth,
		ClosedSpriteY: 0,
		OpenedSpriteX: 12 * config.SpriteWidth,
		OpenedSpriteY: 0,
		// KeyId:         "blue_keycard",
		// Locked:        true,
	})
	e.AddComponent(&component.AppearanceComponent{
		SpriteX:  11 * config.SpriteWidth,
		SpriteY:  0,
		Resource: "map",
	})
	e.AddComponent(&rlcomponents.DescriptionComponent{Name: "Door"})
	l.Level.AddEntity(e)
}

// spawnMaintenanceDoor is the same as spawnDoor but uses maintenance tunnel sprites
// and sets the underlying tile to maintenance tunnel floor.
func spawnMaintenanceDoor(l *world.Level, x, y, z int) {
	l.SetTileTypeAt(x, y, z, world.TypeMaintenanceTunnelFloor)

	e := &ecs.Entity{}
	pc := &rlcomponents.PositionComponent{}
	pc.SetPosition(x, y, z)
	e.AddComponent(pc)
	e.AddComponent(&rlcomponents.SolidComponent{})
	e.AddComponent(&rlcomponents.DoorComponent{
		ClosedSpriteX: 18 * config.SpriteWidth,
		ClosedSpriteY: 0,
		OpenedSpriteX: 18 * config.SpriteWidth,
		OpenedSpriteY: 0,
	})
	e.AddComponent(&component.AppearanceComponent{
		SpriteX:  18 * config.SpriteWidth,
		SpriteY:  0,
		Resource: "map",
	})
	e.AddComponent(&rlcomponents.DescriptionComponent{Name: "Maintenance Door"})
	l.Level.AddEntity(e)
}
