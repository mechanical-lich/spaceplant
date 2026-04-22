package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/stationconfig"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// applyStationConfig stamps config-driven RequiredRoomCounts onto the relevant theme slots.
// FloorStack indices: 0=Engineering, 1=Logistics, 2=Habitation, 3=Commerce, 4=Science, 5=Command.
func applyStationConfig(themes *[]FloorTheme, cfg stationconfig.Config) {
	set := func(idx int, tag string, count int) {
		if idx >= len(*themes) {
			return
		}
		t := &(*themes)[idx]
		if t.RequiredRoomCounts == nil {
			t.RequiredRoomCounts = make(map[string]int)
		}
		if count > 0 {
			t.RequiredRoomCounts[tag] = count
		}
	}

	// Engineering floor (z=0)
	set(0, "engineering_workshop", cfg.EngineeringCapacity)
	set(2, "life_pod_bay", cfg.LifePodBayCount)
	if cfg.SelfDestructEnabled {
		set(5, "self_destruct_room", 1)
	}

	// Habitation floor (z=2)
	set(2, "crew_quarters", cfg.CrewCapacity)

	// Science floor (z=4)
	set(4, "general_lab", cfg.ScienceLabCount)
	set(4, "medical_bay", cfg.MedCount)

	// Command floor (z=5)
	set(5, "security_office", cfg.SecurityCapacity)
}

// FloorResult holds the generated rooms for one floor after generation and population.
type FloorResult struct {
	Z              int
	Theme          *FloorTheme
	Rooms          []Room
	StairX         int // actual X position of the stair tile on this floor
	StairY         int // actual Y position of the stair tile on this floor
	PlacementHints map[int][]PlacementHint // room index → hints from room generators
}

// GenerateFloors generates all floors of the station using the FloorStack theme list.
// It returns a FloorResult per floor so callers can run population passes afterward.
func GenerateFloors(l *world.Level) []FloorResult {
	numFloors := l.Depth
	cfg := stationconfig.Get()

	// Build per-floor theme copies with config-driven RequiredRoomCounts applied.
	themes := make([]FloorTheme, len(FloorStack))
	copy(themes, FloorStack)
	applyStationConfig(&themes, cfg)

	results := make([]FloorResult, numFloors)

	for z := 0; z < numFloors; z++ {
		// Cycle through themes if there are more floors than themes.
		theme := &themes[z%len(themes)]

		rooms := generateFloor(l, z, theme)
		l.Polish(z)

		hints := ApplyRoomGenerators(l, z, rooms)
		l.Polish(z)

		results[z] = FloorResult{
			Z:              z,
			Theme:          theme,
			Rooms:          rooms,
			PlacementHints: hints,
		}
	}

	// Place the main stair column at the guaranteed center, then secondary stairs.
	placeStairs(l, numFloors, results)
	placeSecondaryStairs(l, results)

	// Populate rooms with furniture based on their tags.
	PopulateRooms(l, results)

	if config.Global().DumpGenerationASCII {
		DumpGenerationASCII(l, results)
	}

	return results
}

// placeStairs places the main stair column using two alternating positions so that
// each floor has distinct up and down stair tiles — a "spiral" pattern:
//
//   posA = (cx,   cy)
//   posB = (cx+2, cy)
//
//   z=0 (bottom): StairsUp   at posA
//   z=1:          StairsDown at posA  StairsUp at posB
//   z=2:          StairsDown at posB  StairsUp at posA
//   z=3:          StairsDown at posA  StairsUp at posB
//   ...
//   z=top:        StairsDown at (whichever posA/B matches)
//
// This guarantees you always land on a "down" tile coming from above and a
// distinct "up" tile to continue, with no single tile serving both roles.
func placeStairs(l *world.Level, numFloors int, results []FloorResult) {
	cx := l.Width / 2
	cy := l.Height / 2

	posA := [2]int{cx, cy}
	posB := [2]int{cx + 2, cy}

	// Ensure both positions are floor on every level before placing stair tiles.
	for z := 0; z < numFloors; z++ {
		l.SetTileTypeAt(posA[0], posA[1], z, world.TypeFloor)
		l.SetTileTypeAt(posB[0], posB[1], z, world.TypeFloor)
	}

	for z := 0; z < numFloors; z++ {
		// Which position is the "down" arrival and which is the "up" departure
		// alternates each floor, creating the spiral.
		var downPos, upPos [2]int
		if z%2 == 1 {
			downPos, upPos = posA, posB
		} else {
			downPos, upPos = posB, posA
		}

		if z > 0 {
			l.SetTileTypeAt(downPos[0], downPos[1], z, world.TypeStairsDown)
		}
		if z < numFloors-1 {
			l.SetTileTypeAt(upPos[0], upPos[1], z, world.TypeStairsUp)
		}

		// Record the spawn point as the up-stair tile (where the player arrives
		// from the floor below, which is the previous floor's upPos = this floor's downPos).
		results[z].StairX = downPos[0]
		results[z].StairY = downPos[1]

		l.Polish(z)
	}

	// Floor 0 has no down-stair, so spawn point is the up-stair tile.
	results[0].StairX = posA[0]
	results[0].StairY = posA[1]
}

// placeSecondaryStairs places additional stair pairs between adjacent floors wherever
// the same (x,y) tile is floor on both z and z+1. The number of pairs placed for each
// floor transition is controlled by the lower floor's SecondaryStairCount.
func placeSecondaryStairs(l *world.Level, results []FloorResult) {
	mainX := l.Width / 2
	mainY := l.Height / 2

	for i := 0; i < len(results)-1; i++ {
		lowerZ := results[i].Z
		upperZ := results[i+1].Z
		count := results[i].Theme.SecondaryStairCount
		if count == 0 {
			continue
		}

		// Collect candidate positions: floor on both z levels, not the main column.
		type pos struct{ x, y int }
		var candidates []pos
		for y := 1; y < l.Height-1; y++ {
			for x := 1; x < l.Width-1; x++ {
				// Exclude the center area — the main stair column lives here.
				if absInt(x-mainX) <= 2 && absInt(y-mainY) <= 2 {
					continue
				}
				if l.GetTileType(x, y, lowerZ) == world.TypeFloor &&
					l.GetTileType(x, y, upperZ) == world.TypeFloor {
					candidates = append(candidates, pos{x, y})
				}
			}
		}

		// Shuffle and pick up to count pairs, enforcing minimum spacing so stairs
		// don't cluster together.
		placed := make([]pos, 0, count)
		minSpacing := 10
		for _, c := range candidates {
			if len(placed) >= count {
				break
			}
			tooClose := false
			for _, p := range placed {
				if absInt(c.x-p.x)+absInt(c.y-p.y) < minSpacing {
					tooClose = true
					break
				}
			}
			if tooClose {
				continue
			}
			l.SetTileTypeAt(c.x, c.y, lowerZ, world.TypeStairsUp)
			l.SetTileTypeAt(c.x, c.y, upperZ, world.TypeStairsDown)
			placed = append(placed, c)
		}
	}
}

