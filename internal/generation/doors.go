package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// pendingDoor records a door position to be placed after all carving is done.
type pendingDoor struct {
	x, y, z int
}

// pendingDoors accumulates doors queued during generation.
// Flushed by flushDoors() once all carving for a floor is complete.
var pendingDoors []pendingDoor

// spawnDoor queues a door at (x,y,z) to be placed after all carving is done.
// This prevents later CarveRoom calls from overwriting the door tile.
func spawnDoor(l *world.Level, x, y, z int) {
	pendingDoors = append(pendingDoors, pendingDoor{x, y, z})
}

// flushDoors places all queued doors onto the level and clears the queue.
// Call this after all carving for the floor is complete.
func flushDoors(l *world.Level) {
	for _, d := range pendingDoors {
		tile := l.Level.GetTilePtr(d.x, d.y, d.z)
		if tile == nil {
			continue
		}
		l.SetTileTypeAt(d.x, d.y, d.z, world.TypeFloor)
		e, err := factory.Create("door", d.x, d.y)
		if err != nil {
			continue
		}
		e.GetComponent("Position").(*component.PositionComponent).SetPosition(d.x, d.y, d.z)
		l.Level.AddEntity(e)
	}
	pendingDoors = pendingDoors[:0]
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
