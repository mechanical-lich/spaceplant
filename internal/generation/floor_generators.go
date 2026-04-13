package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// generateFloor dispatches to the correct layout generator for the given theme
// and returns all rooms that were carved, each tagged with a room type.
func generateFloor(l *world.Level, z int, theme *FloorTheme) []Room {
	switch theme.Layout {
	case LayoutRingSpokes:
		return generateRingSpokes(l, z, theme)
	case LayoutGrid:
		return generateGrid(l, z, theme)
	case LayoutIndustrialRing:
		return generateIndustrialRing(l, z, theme)
	case LayoutOpenBays:
		return generateOpenBays(l, z, theme)
	case LayoutRectangle:
		return generateRectangle(l, z, theme)
	default:
		return generateGrid(l, z, theme)
	}
}

// tagRooms assigns a theme-weighted tag to each room in the slice.
func tagRooms(rooms []Room, theme *FloorTheme) []Room {
	for i := range rooms {
		if rooms[i].Tag == "" {
			rooms[i].Tag = theme.pickRoomTag()
		}
	}
	return rooms
}

// generateRingSpokes: circle hub + 4 hallway arms, rooms budded off the arms.
// Good for Command and Science — feels deliberate and spoke-like.
func generateRingSpokes(l *world.Level, z int, theme *FloorTheme) []Room {
	cx := l.Width / 2
	cy := l.Height / 2
	r := l.Width / 8
	if l.Height < l.Width {
		r = l.Height / 8
	}

	// Hub — no-budding so rooms don't grow out of the hub itself
	CarveCircle(l, cx, cy, z, r, world.TypeWall, world.TypeFloor, false, true)

	// Arms
	maxHallW := l.Width/2 - r
	maxHallH := l.Height/2 - r

	// Arms — noOverwrite=false so the arm punches through the circle's wall tile.
	// spawnDoor then converts that shared boundary tile into a passable door.
	hH := utility.GetRandom(5, 10)
	CarveRoom(l, cx+r-1, cy-hH/2, z, maxHallW, hH, world.TypeWall, world.TypeFloor, false, false) // right
	hH = utility.GetRandom(5, 10)
	CarveRoom(l, cx-r+2-maxHallW, cy-hH/2, z, maxHallW, hH, world.TypeWall, world.TypeFloor, false, false) // left
	hW := utility.GetRandom(5, 10)
	CarveRoom(l, cx-hW/2, cy-r+2-maxHallH, z, hW, maxHallH, world.TypeWall, world.TypeFloor, false, false) // up
	hW = utility.GetRandom(5, 10)
	CarveRoom(l, cx-hW/2, cy+r-1, z, hW, maxHallH, world.TypeWall, world.TypeFloor, false, false) // down

	// Doors at the circle boundary — where each arm's first wall tile was carved.
	spawnDoor(l, cx, cy+r-1, z)
	spawnDoor(l, cx, cy-r+1, z)
	spawnDoor(l, cx+r-1, cy, z)
	spawnDoor(l, cx-r+1, cy, z)

	// Bud rooms off the arms
	rooms := BudRooms(l, z, l.Width, l.Height, theme.BudCount)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 15)
	flushDoors(l)

	return tagRooms(rooms, theme)
}

// generateGrid: cross/grid halls with many small budded rooms.
// Good for Habitation and Commerce — lots of small connected rooms.
func generateGrid(l *world.Level, z int, theme *FloorTheme) []Room {
	cx := l.Width / 2
	cy := l.Height / 2

	// Main hall — randomly wide or tall
	wide := utility.GetRandom(0, 2) == 1
	hW := utility.GetRandom(5, 10)
	hH := utility.GetRandom(5, 10)
	if wide {
		hW = utility.GetRandom(l.Width/2, l.Width-l.Width/4)
	} else {
		hH = utility.GetRandom(l.Height/2, l.Height-l.Height/4)
	}
	x1 := cx - hW/2
	y1 := cy - hH/2
	CarveRoom(l, x1, y1, z, hW, hH, world.TypeWall, world.TypeFloor, false, false)

	// Second cross hall (makes a + shape)
	hW2 := utility.GetRandom(5, 10)
	hH2 := utility.GetRandom(5, 10)
	if !wide {
		hW2 = utility.GetRandom(l.Width/2, l.Width-l.Width/4)
	} else {
		hH2 = utility.GetRandom(l.Height/2, l.Height-l.Height/4)
	}
	x2 := cx - hW2/2
	y2 := cy - hH2/2
	CarveRoom(l, x2, y2, z, hW2, hH2, world.TypeWall, world.TypeFloor, true, false)

	// Re-carve both interiors so wall overlap at the intersection doesn't block passage.
	CarveRoom(l, x1+1, y1+1, z, hW-2, hH-2, world.TypeFloor, world.TypeFloor, false, false)
	CarveRoom(l, x2+1, y2+1, z, hW2-2, hH2-2, world.TypeFloor, world.TypeFloor, false, false)

	rooms := BudRooms(l, z, l.Width, l.Height, theme.BudCount)
	l.Polish(z)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 10)
	flushDoors(l)
	l.Polish(z)

	return tagRooms(rooms, theme)
}

// generateIndustrialRing: outer ring + inner ring with maintenance tunnel web.
// Good for Engineering — circular with heavy service access.
func generateIndustrialRing(l *world.Level, z int, theme *FloorTheme) []Room {
	cx := l.Width / 2
	cy := l.Height / 2
	outerR := l.Width / 4
	if l.Height < l.Width {
		outerR = l.Height / 4
	}
	innerR := outerR / 2

	// Outer ring
	CarveCircle(l, cx, cy, z, outerR, world.TypeWall, world.TypeFloor, false, false)
	// Inner ring (hollows out the center, leaving a ring corridor)
	CarveCircle(l, cx, cy, z, innerR, world.TypeWall, world.TypeFloor, false, true)

	// Spoke corridors connecting inner ring wall to outer ring wall.
	// The circle's wall is at distance r-1 from center, so the inner ring wall
	// is at cy+innerR-1 (not cy+innerR). Spokes must start there to overlap it.
	spokeW := utility.GetRandom(3, 5)
	CarveRoom(l, cx-spokeW/2, cy-outerR+1, z, spokeW, outerR-innerR+1, world.TypeWall, world.TypeFloor, false, false) // north spoke
	CarveRoom(l, cx-spokeW/2, cy+innerR-1, z, spokeW, outerR-innerR+1, world.TypeWall, world.TypeFloor, false, false) // south spoke
	CarveRoom(l, cx-outerR+1, cy-spokeW/2, z, outerR-innerR+1, spokeW, world.TypeWall, world.TypeFloor, false, false) // west spoke
	CarveRoom(l, cx+innerR-1, cy-spokeW/2, z, outerR-innerR+1, spokeW, world.TypeWall, world.TypeFloor, false, false) // east spoke

	// Doors on the inner ring wall tiles.
	spawnDoor(l, cx, cy-innerR+1, z)
	spawnDoor(l, cx, cy+innerR-1, z)
	spawnDoor(l, cx-innerR+1, cy, z)
	spawnDoor(l, cx+innerR-1, cy, z)

	rooms := BudRooms(l, z, l.Width, l.Height, theme.BudCount)
	l.Polish(z)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 30)
	flushDoors(l)
	l.Polish(z)

	return tagRooms(rooms, theme)
}

// generateOpenBays: a small number of large rooms, minimal budding.
// Good for Logistics/Cargo — wide open spaces for large operations.
func generateOpenBays(l *world.Level, z int, theme *FloorTheme) []Room {
	var explicitRooms []Room

	// 3–5 large bays placed around the level
	numBays := utility.GetRandom(3, 6)
	bayW := l.Width / 4
	bayH := l.Height / 4

	positions := [][2]int{
		{l.Width/8, l.Height/8},
		{l.Width - l.Width/8 - bayW, l.Height/8},
		{l.Width/8, l.Height - l.Height/8 - bayH},
		{l.Width - l.Width/8 - bayW, l.Height - l.Height/8 - bayH},
		{l.Width/2 - bayW/2, l.Height/2 - bayH/2},
	}

	for i := 0; i < numBays && i < len(positions); i++ {
		w := utility.GetRandom(bayW-4, bayW+4)
		h := utility.GetRandom(bayH-4, bayH+4)
		x, y := positions[i][0], positions[i][1]
		CarveRoom(l, x, y, z, w, h, world.TypeWall, world.TypeFloor, false, false)
		explicitRooms = append(explicitRooms, Room{X: x, Y: y, Width: w, Height: h})
	}

	// Connect bays with wide corridors
	for i := 1; i < len(explicitRooms); i++ {
		a := explicitRooms[i-1]
		b := explicitRooms[i]
		// Horizontal leg
		x1, x2 := a.X+a.Width/2, b.X+b.Width/2
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		CarveRoom(l, x1, a.Y+a.Height/2-2, z, x2-x1, 5, world.TypeWall, world.TypeFloor, false, false)
		// Vertical leg
		y1, y2 := a.Y+a.Height/2, b.Y+b.Height/2
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		CarveRoom(l, b.X+b.Width/2-2, y1, z, 5, y2-y1, world.TypeWall, world.TypeFloor, false, false)
	}

	// Tag the explicit bays
	for i := range explicitRooms {
		explicitRooms[i].Tag = theme.pickRoomTag()
	}

	// A few small budded utility rooms off the corridors
	budded := BudRooms(l, z, l.Width, l.Height, theme.BudCount)
	tagRooms(budded, theme)

	l.Polish(z)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 20)
	flushDoors(l)
	l.Polish(z)

	return append(explicitRooms, budded...)
}

// generateRectangle: corner anchor rooms + connecting halls + central circle.
// Good for a mixed-use floor.
func generateRectangle(l *world.Level, z int, theme *FloorTheme) []Room {
	roomW := l.Width / 6
	roomH := l.Height / 6

	var explicitRooms []Room
	corners := [][2]int{
		{0, 0},
		{l.Width - roomW, 0},
		{0, l.Height - roomH},
		{l.Width - roomW, l.Height - roomH},
	}
	for _, c := range corners {
		CarveRoom(l, c[0], c[1], z, roomW, roomH, world.TypeWall, world.TypeFloor, false, true)
		explicitRooms = append(explicitRooms, Room{X: c[0], Y: c[1], Width: roomW, Height: roomH})
	}

	hallH := 5
	hallW := 5
	// Connecting hallways — noOverwrite=false so they punch through corner room walls.
	CarveRoom(l, roomW-1, roomH/2-hallH/2, z, l.Width-roomW*2+2, hallH, world.TypeWall, world.TypeFloor, false, false)
	CarveRoom(l, roomW-1, l.Height-roomH/2-hallH/2, z, l.Width-roomW*2+2, hallH, world.TypeWall, world.TypeFloor, false, false)
	CarveRoom(l, roomW/2-hallW/2, roomH-1, z, hallW, l.Height-roomH*2+2, world.TypeWall, world.TypeFloor, false, false)
	CarveRoom(l, l.Width-roomW/2-hallW/2, roomH-1, z, hallW, l.Height-roomH*2+2, world.TypeWall, world.TypeFloor, false, false)

	spawnDoor(l, roomW-1, roomH/2, z)
	spawnDoor(l, l.Width-roomW, roomH/2, z)
	spawnDoor(l, roomW-1, l.Height-roomH/2, z)
	spawnDoor(l, l.Width-roomW, l.Height-roomH/2, z)
	spawnDoor(l, roomW/2, roomH-1, z)
	spawnDoor(l, roomW/2, l.Height-roomH, z)
	spawnDoor(l, l.Width-roomW/2, roomH-1, z)
	spawnDoor(l, l.Width-roomW/2, l.Height-roomH, z)

	cx := l.Width / 2
	cy := l.Height / 2
	r := l.Width / 8
	CarveCircle(l, cx, cy, z, r, world.TypeWall, world.TypeFloor, false, true)

	for i := range explicitRooms {
		explicitRooms[i].Tag = theme.pickRoomTag()
	}

	budded := BudRooms(l, z, l.Width, l.Height, theme.BudCount)
	tagRooms(budded, theme)

	l.Polish(z)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 20)
	flushDoors(l)
	l.Polish(z)

	return append(explicitRooms, budded...)
}
