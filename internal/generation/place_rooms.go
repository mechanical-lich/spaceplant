package generation

import (
	"math/rand"

	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type candidate struct {
	wx, wy int // wall tile position
	dx, dy int // outward direction (into open space where the room grows)
}

// PlaceRooms collects all valid room attachment points on the carved skeleton,
// shuffles them, and places up to maxRooms rooms with guaranteed-valid doors.
//
// A valid attachment point is a wall tile that:
//   - has TypeFloor on one cardinal side (hallway side)
//   - has TypeOpen on the opposite side (room grows here)
//   - has TypeWall on both perpendicular sides (not a corner or seam)
//   - is not marked noBudding
//
// The door tile is always the shared boundary wall between the hallway and the
// new room. After flushDoors, both sides of the door will be TypeFloor.
//
// Each room is tagged immediately (via theme.pickRoomTag) so its preferred size
// can be looked up before carving.
func PlaceRooms(l *world.Level, z, maxRooms int, theme *FloorTheme) []Room {
	// --- Candidate collection ---
	var candidates []candidate

	cardinals := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for wy := 1; wy < l.Height-1; wy++ {
		for wx := 1; wx < l.Width-1; wx++ {
			if l.GetTileType(wx, wy, z) != world.TypeWall {
				continue
			}
			if l.GetNoBudding(wx, wy, z) {
				continue
			}

			for _, dir := range cardinals {
				fx, fy := dir[0], dir[1]

				// Hallway side: neighbor in (fx,fy) direction must be floor
				if l.GetTileType(wx+fx, wy+fy, z) != world.TypeFloor {
					continue
				}

				// Outward side: neighbor in opposite direction must be open
				ox, oy := -fx, -fy
				if l.GetTileType(wx+ox, wy+oy, z) != world.TypeOpen {
					continue
				}

				// Perpendicular neighbors must both be wall (no corners/seams)
				var perpA, perpB [2]int
				if fx == 0 {
					perpA = [2]int{wx - 1, wy}
					perpB = [2]int{wx + 1, wy}
				} else {
					perpA = [2]int{wx, wy - 1}
					perpB = [2]int{wx, wy + 1}
				}
				if l.GetTileType(perpA[0], perpA[1], z) != world.TypeWall {
					continue
				}
				if l.GetTileType(perpB[0], perpB[1], z) != world.TypeWall {
					continue
				}

				candidates = append(candidates, candidate{wx, wy, ox, oy})
				break // one direction per wall tile
			}
		}
	}

	// --- Shuffle ---
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// --- Room placement ---
	var rooms []Room
	placed := 0

	// Build a queue: required rooms first, then random weighted tags.
	type taggedCandidate struct {
		c   candidate
		tag string
	}
	var queue []taggedCandidate

	// Build the full required-tag list: RequiredRooms (×1 each) + RequiredRoomCounts (×N each).
	var requiredTags []string
	if theme != nil {
		requiredTags = append(requiredTags, theme.RequiredRooms...)
		for tag, count := range theme.RequiredRoomCounts {
			for i := 0; i < count; i++ {
				requiredTags = append(requiredTags, tag)
			}
		}
	}

	// Assign required tags to the first N candidates.
	ci := 0
	if theme != nil {
		for _, tag := range requiredTags {
			for ci < len(candidates) {
				c := candidates[ci]
				ci++
				sz := RoomSizeFor(tag)
				rW := utility.GetRandom(sz.MinW, sz.MaxW+1)
				rH := utility.GetRandom(sz.MinH, sz.MaxH+1)
				var rx, ry int
				switch {
				case c.dx == 0 && c.dy == -1:
					rx, ry = c.wx-rW/2, c.wy-rH+1
				case c.dx == 0 && c.dy == 1:
					rx, ry = c.wx-rW/2, c.wy
				case c.dx == -1 && c.dy == 0:
					rx, ry = c.wx-rW+1, c.wy-rH/2
				default:
					rx, ry = c.wx, c.wy-rH/2
				}
				if rx < 0 || ry < 0 || rx+rW > l.Width || ry+rH > l.Height {
					continue
				}
				var intersects bool
				switch {
				case c.dx == 0 && c.dy == -1:
					intersects = RoomIntersects(l, z, rx, ry, rW, rH-1)
				case c.dx == 0 && c.dy == 1:
					intersects = RoomIntersects(l, z, rx, ry+1, rW, rH-1)
				case c.dx == -1 && c.dy == 0:
					intersects = RoomIntersects(l, z, rx, ry, rW-1, rH)
				default:
					intersects = RoomIntersects(l, z, rx+1, ry, rW-1, rH)
				}
				if intersects {
					continue
				}
				CarveRoom(l, rx, ry, z, rW, rH, world.TypeWall, world.TypeFloor, true, false)
				spawnDoor(l, c.wx, c.wy, z)
				rooms = append(rooms, Room{X: rx, Y: ry, Width: rW, Height: rH, Tag: tag, DoorDir: [2]int{c.dx, c.dy}})
				placed++
				break
			}
		}
	}

	// Remaining candidates get random weighted tags.
	for _, c := range candidates[ci:] {
		queue = append(queue, taggedCandidate{c: c})
	}

	for _, tc := range queue {
		if placed >= maxRooms {
			break
		}
		c := tc.c

		var tag string
		if theme != nil {
			tag = theme.pickRoomTag()
		}
		sz := RoomSizeFor(tag)
		rW := utility.GetRandom(sz.MinW, sz.MaxW+1)
		rH := utility.GetRandom(sz.MinH, sz.MaxH+1)

		// Compute room origin so the door tile (c.wx, c.wy) sits on the room border.
		var rx, ry int
		switch {
		case c.dx == 0 && c.dy == -1: // North — room grows up
			rx = c.wx - rW/2
			ry = c.wy - rH + 1
		case c.dx == 0 && c.dy == 1: // South — room grows down
			rx = c.wx - rW/2
			ry = c.wy
		case c.dx == -1 && c.dy == 0: // West — room grows left
			rx = c.wx - rW + 1
			ry = c.wy - rH/2
		default: // East — room grows right
			rx = c.wx
			ry = c.wy - rH/2
		}

		// Bounds check
		if rx < 0 || ry < 0 || rx+rW > l.Width || ry+rH > l.Height {
			continue
		}

		// Intersection check — exclude the shared wall row/col (door tile is TypeWall,
		// including it would always fail the check).
		var intersects bool
		switch {
		case c.dx == 0 && c.dy == -1: // North: exclude bottom row (door row)
			intersects = RoomIntersects(l, z, rx, ry, rW, rH-1)
		case c.dx == 0 && c.dy == 1: // South: exclude top row
			intersects = RoomIntersects(l, z, rx, ry+1, rW, rH-1)
		case c.dx == -1 && c.dy == 0: // West: exclude rightmost col
			intersects = RoomIntersects(l, z, rx, ry, rW-1, rH)
		default: // East: exclude leftmost col
			intersects = RoomIntersects(l, z, rx+1, ry, rW-1, rH)
		}
		if intersects {
			continue
		}

		CarveRoom(l, rx, ry, z, rW, rH, world.TypeWall, world.TypeFloor, true, false)
		spawnDoor(l, c.wx, c.wy, z)
		rooms = append(rooms, Room{X: rx, Y: ry, Width: rW, Height: rH, Tag: tag, DoorDir: [2]int{c.dx, c.dy}})
		placed++
	}

	return rooms
}
