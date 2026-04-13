package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// FloorResult holds the generated rooms for one floor after generation and population.
type FloorResult struct {
	Z     int
	Theme *FloorTheme
	Rooms []Room
}

// GenerateFloors generates all floors of the station using the FloorStack theme list.
// It returns a FloorResult per floor so callers can run population passes afterward.
// Stairs connecting adjacent floors are placed at (stairX, stairY).
func GenerateFloors(l *world.Level, stairX, stairY int) []FloorResult {
	numFloors := l.Depth
	themes := FloorStack

	results := make([]FloorResult, numFloors)

	for z := 0; z < numFloors; z++ {
		// Cycle through themes if there are more floors than themes.
		theme := &themes[z%len(themes)]

		rooms := generateFloor(l, z, theme)
		l.Polish(z)

		results[z] = FloorResult{
			Z:     z,
			Theme: theme,
			Rooms: rooms,
		}
	}

	// Place stairs connecting each floor to the next.
	placeStairs(l, stairX, stairY, numFloors)

	// Populate rooms with furniture based on their tags.
	PopulateRooms(l, results)

	return results
}

// placeStairs sets stair tiles on each floor and polishes afterward.
// Floor 0 gets only a StairsUp; the top floor gets only a StairsDown;
// middle floors get both.
func placeStairs(l *world.Level, x, y, numFloors int) {
	for z := 0; z < numFloors; z++ {
		// Ensure the stair tile is on walkable ground.
		// If the chosen position is solid or open space, nudge to a nearby floor tile.
		sx, sy := findNearbyFloor(l, x, y, z)

		if z < numFloors-1 {
			l.SetTileTypeAt(sx, sy, z, world.TypeStairsUp)
		}
		if z > 0 {
			// Place the down-stair on the same XY on the floor above.
			l.SetTileTypeAt(sx, sy, z, world.TypeStairsDown)
		}
		l.Polish(z)
	}
}

// findNearbyFloor returns the closest walkable floor tile to (x, y) on layer z,
// searching in an expanding square. Falls back to (x, y) if nothing found.
func findNearbyFloor(l *world.Level, x, y, z int) (int, int) {
	for radius := 0; radius <= 10; radius++ {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				nx, ny := x+dx, y+dy
				if nx < 0 || ny < 0 || nx >= l.Width || ny >= l.Height {
					continue
				}
				t := l.Level.GetTilePtr(nx, ny, z)
				if t != nil && t.Type == world.TypeFloor {
					return nx, ny
				}
			}
		}
	}
	return x, y
}
