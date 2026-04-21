package scenario

import (
	"strconv"

	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// FloorMatches reports whether floor z with the given theme name satisfies rule.Floors.
// An empty Floors list matches any floor.
func (r SpawnRule) FloorMatches(z int, themeName string) bool {
	if len(r.Floors) == 0 {
		return true
	}
	idxStr := strconv.Itoa(z)
	for _, f := range r.Floors {
		if f == themeName || f == idxStr {
			return true
		}
	}
	return false
}

// SpawnTiles returns all valid (x, y) positions for a blueprint on floor z.
// If rule.Rooms is non-empty, only tiles inside matching rooms are considered.
// If rule.Rooms is empty, all passable non-open tiles on the floor qualify.
func SpawnTiles(l *world.Level, z int, fr generation.FloorResult, rule SpawnRule) [][2]int {
	if len(rule.Rooms) == 0 {
		return anyTiles(l, z)
	}

	roomSet := make(map[string]bool, len(rule.Rooms))
	for _, tag := range rule.Rooms {
		roomSet[tag] = true
	}

	var tiles [][2]int
	for _, room := range fr.Rooms {
		if !roomSet[room.Tag] {
			continue
		}
		for rx := room.X; rx < room.X+room.Width; rx++ {
			for ry := room.Y; ry < room.Y+room.Height; ry++ {
				t := l.Level.GetTilePtr(rx, ry, z)
				if t == nil || t.IsSolid() || t.Type == world.TypeOpen {
					continue
				}
				tiles = append(tiles, [2]int{rx, ry})
			}
		}
	}
	return tiles
}

// FilterByFloor returns the subset of blueprints whose SpawnRule allows floor z.
func FilterByFloor(blueprints []string, rules map[string]SpawnRule, z int, themeName string) []string {
	out := make([]string, 0, len(blueprints))
	for _, bp := range blueprints {
		rule := rules[bp]
		if rule.FloorMatches(z, themeName) {
			out = append(out, bp)
		}
	}
	return out
}

// anyTiles collects all passable, non-open tiles on floor z.
func anyTiles(l *world.Level, z int) [][2]int {
	var tiles [][2]int
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			t := l.Level.GetTilePtr(x, y, z)
			if t == nil || t.IsSolid() || t.Type == world.TypeOpen {
				continue
			}
			tiles = append(tiles, [2]int{x, y})
		}
	}
	return tiles
}
