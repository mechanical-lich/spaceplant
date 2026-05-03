package generation

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/path"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

func GenerateRoundStation(l *world.Level, z int) {
	// Center point
	x := l.Width / 2
	y := l.Height / 2

	// Center is 1/4 the width (or height)
	r := l.Width / 8
	if l.Height > l.Width {
		r = l.Height / 8
	}

	// Carve central room
	CarveCircle(l, x, y, z, r, world.TypeWall, world.TypeFloor, false, true)   // Outer
	CarveCircle(l, x, y, z, r/2, world.TypeWall, world.TypeFloor, false, true) // Inner

	// Hallways — carve before placing doors so hallway borders don't overwrite door tiles
	//Right
	maxHallWidth := l.Width/2 - r
	hWidth := maxHallWidth
	hHeight := rlmath.GetRandom(5, 10)
	CarveRoom(l, x+r-1, y-hHeight/2, z, hWidth, hHeight, world.TypeWall, world.TypeFloor, true, false)

	//Left
	hWidth = maxHallWidth
	hHeight = rlmath.GetRandom(5, 10)
	CarveRoom(l, x-r+2-hWidth, y-hHeight/2, z, hWidth, hHeight, world.TypeWall, world.TypeFloor, true, false)

	//Up
	maxHallHeight := l.Height/2 - r
	hHeight = maxHallHeight
	hWidth = rlmath.GetRandom(5, 10)
	CarveRoom(l, x-hWidth/2, y-r+2-hHeight, z, hWidth, hHeight, world.TypeWall, world.TypeFloor, true, false)

	//Down
	hHeight = maxHallHeight
	hWidth = rlmath.GetRandom(5, 10)
	CarveRoom(l, x-hWidth/2, y+r-1, z, hWidth, hHeight, world.TypeWall, world.TypeFloor, true, false)

	// Rooms and tunnels
	PlaceRooms(l, z, 100, nil)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 30)

	// Doors last — after all carving so no subsequent CarveRoom overwrites the floor tile
	// Inner Doors
	spawnDoor(l, x, y+r/2-1, z)
	spawnDoor(l, x, y-r/2+1, z)
	spawnDoor(l, x+r/2-1, y, z)
	spawnDoor(l, x-r/2+1, y, z)

	// Outer Doors
	spawnDoor(l, x, y+r-1, z)
	spawnDoor(l, x, y-r+1, z)
	spawnDoor(l, x+r-1, y, z)
	spawnDoor(l, x-r+1, y, z)
}

func GenerateRectangleStation(l *world.Level, z int) {
	roomWidth := l.Width / 6
	roomHeight := l.Height / 6

	//Top Left
	CarveRoom(l, 0, 0, z, roomWidth, roomHeight, world.TypeWall, world.TypeFloor, false, true)

	//Top Right
	CarveRoom(l, l.Width-roomWidth, 0, z, roomWidth, roomHeight, world.TypeWall, world.TypeFloor, false, true)

	//Bottom Left
	CarveRoom(l, 0, l.Height-roomHeight, z, roomWidth, roomHeight, world.TypeWall, world.TypeFloor, false, true)

	//Bottom right
	CarveRoom(l, l.Width-roomWidth, l.Height-roomHeight, z, roomWidth, roomHeight, world.TypeWall, world.TypeFloor, false, true)

	// Hallways
	hallwayHeight := 5
	hallwayWidth := 5
	//Top
	CarveRoom(l, roomWidth-1, roomHeight/2-hallwayHeight/2, z, l.Width-roomWidth*2+2, hallwayHeight, world.TypeWall, world.TypeFloor, true, false)
	//Bottom
	CarveRoom(l, roomWidth-1, l.Height-roomHeight/2-hallwayHeight/2, z, l.Width-roomWidth*2+2, hallwayHeight, world.TypeWall, world.TypeFloor, true, false)
	//Left
	CarveRoom(l, roomWidth/2-hallwayWidth/2, roomHeight-1, z, hallwayWidth, l.Height-roomHeight*2+2, world.TypeWall, world.TypeFloor, true, false)
	//Right
	CarveRoom(l, l.Width-roomWidth/2-hallwayWidth/2, roomHeight-1, z, hallwayWidth, l.Height-roomHeight*2+2, world.TypeWall, world.TypeFloor, true, false)

	// Main hallway doors
	spawnDoor(l, roomWidth-1, roomHeight/2, z)
	spawnDoor(l, l.Width-roomWidth, roomHeight/2, z)
	spawnDoor(l, roomWidth-1, l.Height-roomHeight/2, z)
	spawnDoor(l, l.Width-roomWidth, l.Height-roomHeight/2, z)
	spawnDoor(l, roomWidth/2, roomHeight-1, z)
	spawnDoor(l, roomWidth/2, l.Height-roomHeight, z)
	spawnDoor(l, l.Width-roomWidth/2, roomHeight-1, z)
	spawnDoor(l, l.Width-roomWidth/2, l.Height-roomHeight, z)

	// Central circle
	x := l.Width / 2
	y := l.Height / 2
	r := l.Width / 8
	CarveCircle(l, x, y, z, r, world.TypeWall, world.TypeFloor, false, true)

	// Central tunnels
	CarveMaintenanceTunnel(l, z, l.Width/2, roomHeight/2+hallwayHeight/2, x, y-r+1, world.TypeMaintenanceTunnelFloor)
	CarveMaintenanceTunnel(l, z, l.Width/2, l.Height-roomHeight/2-hallwayHeight/2, x, y+r-1, world.TypeMaintenanceTunnelFloor)
	CarveMaintenanceTunnel(l, z, roomWidth/2+hallwayWidth/2, l.Height/2, x-r+1, y, world.TypeMaintenanceTunnelFloor)
	CarveMaintenanceTunnel(l, z, l.Width-roomWidth/2-hallwayWidth/2, l.Height/2, x+r-1, y, world.TypeMaintenanceTunnelFloor)

	l.Polish(z)
	PlaceRooms(l, z, 50, nil)

	l.Polish(z)
	CarveMaintenanceTunnels(l, z, l.Width, l.Height, 30)

	l.Polish(z)
}

func GenerateStation(l *world.Level, z, width, height int) {
	// Create center room
	x := width / 2
	y := height / 2

	// Generate the central hallway
	wide := rlmath.GetRandom(0, 2) == 1
	tWidth := rlmath.GetRandom(5, 10)
	tHeight := rlmath.GetRandom(5, 10)
	if wide {
		tWidth = rlmath.GetRandom(width/2, width-width/4)
	} else {
		tHeight = rlmath.GetRandom(height/2, height-height/4)
	}

	x -= tWidth / 2
	y -= tHeight / 2
	CarveRoom(l, x, y, z, tWidth, tHeight, world.TypeWall, world.TypeFloor, false, false)

	// Optional second hallway
	if rlmath.GetRandom(0, 10) > 5 {
		x2 := width / 2
		y2 := height / 2
		tWidth2 := rlmath.GetRandom(5, 10)
		tHeight2 := rlmath.GetRandom(5, 10)
		// Makes a +
		if !wide {
			tWidth2 = rlmath.GetRandom(width/2, width-width/4)
		} else {
			tHeight2 = rlmath.GetRandom(height/2, height-height/4)
		}

		x2 -= tWidth2 / 2
		y2 -= tHeight2 / 2
		CarveRoom(l, x2, y2, z, tWidth2, tHeight2, world.TypeWall, world.TypeFloor, true, false)
		CarveRoom(l, x+1, y+1, z, tWidth-2, tHeight-2, world.TypeFloor, world.TypeFloor, false, false)
		CarveRoom(l, x2+1, y2+1, z, tWidth2-2, tHeight2-2, world.TypeFloor, world.TypeFloor, false, false)
	}

	PlaceRooms(l, z, 50, nil)

	// Polish so we can pathfind
	l.Polish(z)

	// Maintenance Tunnels
	CarveMaintenanceTunnels(l, z, width, height, 10)

	l.Polish(z)
}

func CarveMaintenanceTunnels(l *world.Level, z, width, height, numTunnels int) {
	for i := 0; i < numTunnels; i++ {
		done := false
		tries := 0
		for tries < 99999 && !done {
			tries++
			tX1 := rlmath.GetRandom(0, width)
			tY1 := rlmath.GetRandom(0, height)

			if l.GetTileType(tX1, tY1, z) == world.TypeWall {
				if !l.TileNeighbors(tX1, tY1, z, world.TypeOpen) || !l.TileNeighbors(tX1, tY1, z, world.TypeFloor) {
					continue
				}
			} else {
				continue
			}

			tX2 := rlmath.GetRandom(0, width)
			tY2 := rlmath.GetRandom(0, height)

			if tX2 == tX1 && tY2 == tY1 {
				continue
			}

			if l.GetTileType(tX2, tY2, z) == world.TypeWall {
				if !l.TileNeighbors(tX2, tY2, z, world.TypeOpen) || !l.TileNeighbors(tX2, tY2, z, world.TypeFloor) {
					continue
				}
			} else {
				continue
			}

			// Temporarily make endpoints non-solid for pathfinding by setting to floor
			t1 := l.Level.GetTilePtr(tX1, tY1, z)
			t2 := l.Level.GetTilePtr(tX2, tY2, z)
			origType1 := t1.Type
			origType2 := t2.Type
			t1.Type = world.TypeFloor
			t2.Type = world.TypeFloor

			steps, distance, success := path.Path(l.Level, t1.Idx, t2.Idx)

			// Restore original types
			t1.Type = origType1
			t2.Type = origType2

			if !success || distance > 10 || distance < 2 {
				continue
			}

			// Reject paths that pass through existing rooms — tunnels must only
			// travel through open (void) space between structures.
			throughRoom := false
			for i, stepID := range steps {
				if i == 0 || i == len(steps)-1 {
					continue // endpoints are boundary wall tiles, always ok
				}
				t := l.Level.GetTilePtrIndex(stepID)
				if t.Type != world.TypeOpen {
					throughRoom = true
					break
				}
			}
			if throughRoom {
				continue
			}

			for _, stepID := range steps {
				t := l.Level.GetTilePtrIndex(stepID)
				tx, ty, _ := t.Coords()
				l.SetTileTypeAt(tx, ty, z, world.TypeMaintenanceTunnelFloor)
			}

			spawnMaintenanceDoor(l, tX1, tY1, z)
			spawnMaintenanceDoor(l, tX2, tY2, z)
			done = true
		}
	}
}

func CarveMaintenanceTunnel(l *world.Level, z, x1, y1, x2, y2, floor int) bool {
	t1 := l.Level.GetTilePtr(x1, y1, z)
	t2 := l.Level.GetTilePtr(x2, y2, z)
	origType1 := t1.Type
	origType2 := t2.Type
	t1.Type = world.TypeFloor
	t2.Type = world.TypeFloor

	steps, _, success := path.Path(l.Level, t1.Idx, t2.Idx)

	t1.Type = origType1
	t2.Type = origType2

	if !success {
		return false
	}

	// Reject paths that pass through existing rooms.
	for i, stepID := range steps {
		if i == 0 || i == len(steps)-1 {
			continue
		}
		t := l.Level.GetTilePtrIndex(stepID)
		if t.Type != world.TypeOpen {
			return false
		}
	}

	for _, stepID := range steps {
		t := l.Level.GetTilePtrIndex(stepID)
		tx, ty, _ := t.Coords()
		l.SetTileTypeAt(tx, ty, z, floor)
	}

	spawnMaintenanceDoor(l, x1, y1, z)
	spawnMaintenanceDoor(l, x2, y2, z)

	return true
}

// findNearestFloor returns the coordinates of the nearest TypeFloor tile to (cx,cy)
// by expanding outward up to maxDist tiles (Manhattan-distance BFS).
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func findNearestFloor(l *world.Level, cx, cy, z, maxDist int) (int, int, bool) {
	for d := 0; d <= maxDist; d++ {
		for dx := -d; dx <= d; dx++ {
			for _, dy := range []int{d - absInt(dx), -(d - absInt(dx))} {
				x, y := cx+dx, cy+dy
				if x < 0 || x >= l.Width || y < 0 || y >= l.Height {
					continue
				}
				if l.GetTileType(x, y, z) == world.TypeFloor {
					return x, y, true
				}
				if dy == 0 {
					break // avoid visiting (cx+dx, cy+0) twice
				}
			}
		}
	}
	return 0, 0, false
}

// ConnectDisconnectedRegions finds floor tiles unreachable from (startX, startY)
// and carves maintenance tunnels to connect them. Returns the number of tunnels carved.
func ConnectDisconnectedRegions(l *world.Level, startX, startY, z int) int {
	// Find a valid floor start — walk outward if the given point isn't floor.
	sx, sy, ok := findNearestFloor(l, startX, startY, z, 20)
	if !ok {
		return 0
	}

	tunnels := 0
	for iter := 0; iter < 20; iter++ {
		reachable := FloodFillReachable(l, sx, sy, z, nil)
		reachSet := make(map[[2]int]bool, len(reachable))
		for _, c := range reachable {
			reachSet[c] = true
		}

		// Collect one tile per isolated floor region.
		var isolated [][2]int
		for y := 0; y < l.Height; y++ {
			for x := 0; x < l.Width; x++ {
				tt := l.GetTileType(x, y, z)
				if tt != world.TypeFloor && tt != world.TypeMaintenanceTunnelFloor {
					continue
				}
				if !reachSet[[2]int{x, y}] {
					isolated = append(isolated, [2]int{x, y})
					break // one representative per scan is enough; re-flood after each tunnel
				}
			}
			if len(isolated) > 0 {
				break
			}
		}

		if len(isolated) == 0 {
			break
		}

		target := isolated[0]

		// Find the nearest reachable tile to the target.
		bestDist := 1<<31 - 1
		var bestR [2]int
		for _, r := range reachable {
			d := absInt(r[0]-target[0]) + absInt(r[1]-target[1])
			if d < bestDist {
				bestDist = d
				bestR = r
			}
		}

		CarveMaintenanceTunnel(l, z, bestR[0], bestR[1], target[0], target[1], world.TypeMaintenanceTunnelFloor)
		tunnels++
	}
	return tunnels
}

func RoomIntersects(l *world.Level, z, x, y, width, height int) bool {
	if x+width > l.Width || y+height > l.Height {
		return true
	}

	if y < 0 || x < 0 {
		return true
	}

	for currentY := 0; currentY < height; currentY++ {
		for currentX := 0; currentX < width; currentX++ {
			t := l.Level.GetTilePtr(currentX+x, currentY+y, z)
			if t == nil {
				return true
			}
			if t.Type != world.TypeOpen {
				return true
			}
		}
	}

	return false
}
