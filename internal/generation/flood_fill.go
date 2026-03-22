package generation

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// FloodFillReachable returns all tile coordinates reachable from (startX, startY)
// on the given Z layer via BFS. Door entities in blockedDoors are treated as
// impassable walls; all other closed-but-unlocked doors and solid non-door entities
// (creatures etc.) are treated as passable, since the player can open or fight past them.
func FloodFillReachable(
	l *world.Level,
	startX, startY, z int,
	blockedDoors map[*ecs.Entity]bool,
) [][2]int {
	type coord struct{ x, y int }

	visited := make(map[int]bool)
	queue := []coord{{startX, startY}}

	idx := func(x, y int) int { return x + y*l.Width }

	visited[idx(startX, startY)] = true
	var result [][2]int

	offsets := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		result = append(result, [2]int{cur.x, cur.y})

		for _, off := range offsets {
			nx, ny := cur.x+off[0], cur.y+off[1]
			if nx < 0 || nx >= l.Width || ny < 0 || ny >= l.Height {
				continue
			}
			if visited[idx(nx, ny)] {
				continue
			}
			tile := l.Level.GetTilePtr(nx, ny, z)
			if tile == nil || tile.IsSolid() {
				continue
			}
			// Check for a door entity that is explicitly blocked.
			if ent := l.Level.GetSolidEntityAt(nx, ny, z); ent != nil {
				if ent.HasComponent(rlcomponents.Door) && blockedDoors[ent] {
					continue
				}
				// Non-door solid entities (creatures) are treated as passable.
			}
			visited[idx(nx, ny)] = true
			queue = append(queue, coord{nx, ny})
		}
	}

	return result
}
