package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// spawnDoor replaces a tile door placement with a door entity.
// The underlying tile is set to floor so the entity provides all blocking/visual.
func spawnDoor(l *world.Level, x, y, z int) {
	l.SetTileTypeAt(x, y, z, world.TypeFloor)
	e, err := factory.Create("door", x, y)
	if err != nil {
		panic("spawnDoor: " + err.Error())
	}
	e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
	l.Level.AddEntity(e)
}

// spawnMaintenanceDoor is like spawnDoor but uses maintenance tunnel sprites
// and sets the underlying tile to maintenance tunnel floor.
func spawnMaintenanceDoor(l *world.Level, x, y, z int) {
	l.SetTileTypeAt(x, y, z, world.TypeMaintenanceTunnelFloor)
	e, err := factory.Create("maintenance_door", x, y)
	if err != nil {
		panic("spawnMaintenanceDoor: " + err.Error())
	}
	e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
	l.Level.AddEntity(e)
}
