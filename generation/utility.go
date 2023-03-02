package generation

import (
	"github.com/beefsack/go-astar"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/utility"
)

func GenerateRoundStation(l *level.Level) {
	// Center point
	x := l.Width / 2
	y := l.Height / 2

	// Center is 1/4 the width (or height)
	r := l.Width / 8
	if l.Height > l.Width {
		r = l.Height / 8
	}

	// Carve central room
	CarveCircle(l, x, y, r, level.Type_Wall, level.Type_Floor, false, true)   // Outer
	CarveCircle(l, x, y, r/2, level.Type_Wall, level.Type_Floor, false, true) // Inner

	// TODO better this.   If the width or height changes then the doors might be placed wrong
	// Inner Doors
	l.SetTileType(x, y+r/2-1, level.Type_Door)
	l.SetTileType(x, y-r/2+1, level.Type_Door)
	l.SetTileType(x+r/2-1, y, level.Type_Door)
	l.SetTileType(x-r/2+1, y, level.Type_Door)

	// Outer Doors
	l.SetTileType(x, y+r-1, level.Type_Door)
	l.SetTileType(x, y-r+1, level.Type_Door)
	l.SetTileType(x+r-1, y, level.Type_Door)
	l.SetTileType(x-r+1, y, level.Type_Door)

	// Hallways

	//Right
	maxHallWidth := l.Width/2 - r
	hWidth := maxHallWidth //utility.GetRandom(maxHallWidth/2, maxHallWidth)
	hHeight := utility.GetRandom(5, 10)
	CarveRoom(l, x+r-1, y-hHeight/2, hWidth, hHeight, level.Type_Wall, level.Type_Floor, true, false)

	//Left
	hWidth = maxHallWidth //utility.GetRandom(maxHallWidth/2, maxHallWidth)
	hHeight = utility.GetRandom(5, 10)
	CarveRoom(l, x-r+2-hWidth, y-hHeight/2, hWidth, hHeight, level.Type_Wall, level.Type_Floor, true, false)

	//Up
	maxHallHeight := l.Height/2 - r
	hHeight = maxHallHeight //utility.GetRandom(maxHallHeight/2, maxHallHeight)
	hWidth = utility.GetRandom(5, 10)
	CarveRoom(l, x-hWidth/2, y-r+2-hHeight, hWidth, hHeight, level.Type_Wall, level.Type_Floor, true, false)

	//Down
	hHeight = maxHallHeight //utility.GetRandom(maxHallHeight/2, maxHallHeight)
	hWidth = utility.GetRandom(5, 10)
	CarveRoom(l, x-hWidth/2, y+r-1, hWidth, hHeight, level.Type_Wall, level.Type_Floor, true, false)

	// Rooms and tunnels
	BudRooms(l, l.Width, l.Height, 100)
	CarveMaintenanceTunnels(l, l.Width, l.Height, 30)

}

func GenerateRectangleStation(l *level.Level) {
	roomWidth := l.Width / 6
	roomHeight := l.Height / 6

	//Top Left
	CarveRoom(l, 0, 0, roomWidth, roomHeight, level.Type_Wall, level.Type_Floor, false, true)

	//Top Right
	CarveRoom(l, l.Width-roomWidth, 0, roomWidth, roomHeight, level.Type_Wall, level.Type_Floor, false, true)

	//Bottom Left
	CarveRoom(l, 0, l.Height-roomHeight, roomWidth, roomHeight, level.Type_Wall, level.Type_Floor, false, true)

	//Bottom right
	CarveRoom(l, l.Width-roomWidth, l.Height-roomHeight, roomWidth, roomHeight, level.Type_Wall, level.Type_Floor, false, true)

	// Hallways
	//Top
	hallwayHeight := 5
	hallwayWidth := 5
	CarveRoom(l, roomWidth-1, roomHeight/2-hallwayHeight/2, l.Width-roomWidth*2+2, hallwayHeight, level.Type_Wall, level.Type_Floor, true, false)

	//Bottom
	CarveRoom(l, roomWidth-1, l.Height-roomHeight/2-hallwayHeight/2, l.Width-roomWidth*2+2, hallwayHeight, level.Type_Wall, level.Type_Floor, true, false)

	//Left
	CarveRoom(l, roomWidth/2-hallwayWidth/2, roomHeight-1, hallwayWidth, l.Height-roomHeight*2+2, level.Type_Wall, level.Type_Floor, true, false)

	//Right
	CarveRoom(l, l.Width-roomWidth/2-hallwayWidth/2, roomHeight-1, hallwayWidth, l.Height-roomHeight*2+2, level.Type_Wall, level.Type_Floor, true, false)

	// Main hallway doors
	//Top
	l.SetTileType(roomWidth-1, roomHeight/2, level.Type_Door)
	l.SetTileType(l.Width-roomWidth, roomHeight/2, level.Type_Door)
	//Bottom
	l.SetTileType(roomWidth-1, l.Height-roomHeight/2, level.Type_Door)
	l.SetTileType(l.Width-roomWidth, l.Height-roomHeight/2, level.Type_Door)

	//Left
	l.SetTileType(roomWidth/2, roomHeight-1, level.Type_Door)
	l.SetTileType(roomWidth/2, l.Height-roomHeight, level.Type_Door)

	//Right
	l.SetTileType(l.Width-roomWidth/2, roomHeight-1, level.Type_Door)
	l.SetTileType(l.Width-roomWidth/2, l.Height-roomHeight, level.Type_Door)

	// Central circle
	x := l.Width / 2
	y := l.Height / 2

	r := l.Width / 8
	CarveCircle(l, x, y, r, level.Type_Wall, level.Type_Floor, false, true)

	// Central tunnels
	//Top
	CarveMaintenanceTunnel(l, l.Width/2, roomHeight/2+hallwayHeight/2, x, y-r+1, level.Type_MaintenanceTunnelFLoor, level.Type_MaintenanceTunnelDoor)
	//Bottom
	CarveMaintenanceTunnel(l, l.Width/2, l.Height-roomHeight/2-hallwayHeight/2, x, y+r-1, level.Type_MaintenanceTunnelFLoor, level.Type_MaintenanceTunnelDoor)
	//Left
	CarveMaintenanceTunnel(l, roomWidth/2+hallwayWidth/2, l.Height/2, x-r+1, y, level.Type_MaintenanceTunnelFLoor, level.Type_MaintenanceTunnelDoor)
	//Right
	CarveMaintenanceTunnel(l, l.Width-roomWidth/2-hallwayWidth/2, l.Height/2, x+r-1, y, level.Type_MaintenanceTunnelFLoor, level.Type_MaintenanceTunnelDoor)

	l.Polish()
	// Rooms and tunnels
	BudRooms(l, l.Width, l.Height, 50)

	l.Polish()
	CarveMaintenanceTunnels(l, l.Width, l.Height, 30)

	l.Polish()
}

func GenerateStation(l *level.Level, width, height int) {
	// Create center room
	x := width / 2
	y := height / 2

	//
	// Generate the central hallway
	//
	wide := utility.GetRandom(0, 2) == 1
	tWidth := utility.GetRandom(5, 10)
	tHeight := utility.GetRandom(5, 10)
	if wide {
		tWidth = utility.GetRandom(width/2, width-width/4)
	} else {
		tHeight = utility.GetRandom(height/2, height-height/4)
	}

	x -= tWidth / 2
	y -= tHeight / 2
	CarveRoom(l, x, y, tWidth, tHeight, level.Type_Wall, level.Type_Floor, false, false)

	// //
	// // End cap room 1
	// //
	// eX := x + tWidth - 1
	// eY := y + tHeight - 1
	// eWidth := utility.GetRandom(tWidth+2, tWidth+10)
	// eHeight := utility.GetRandom(tHeight+2, tHeight+10)

	// dX := eX
	// dY := eY
	// sX := eX
	// sY := eY
	// if wide {
	// 	//eWidth = width - tWidth - 1
	// 	eY = y - tHeight/4
	// 	eWidth = utility.GetRandom(5, 10)
	// 	dY -= tHeight / 2
	// 	sX -= tWidth
	// } else {
	// 	//eHeight = height - tHeight - 1
	// 	eX = x - tWidth/4
	// 	eHeight = utility.GetRandom(5, 10)
	// 	dX -= tWidth / 2
	// 	sY -= tHeight
	// }

	// CarveRoom(l, eX, eY, eWidth, eHeight, level.Type_Wall, level.Type_Floor, true, true)
	// l.SetTileType(dX, dY, level.Type_Door)
	// l.SetTileType(sX, sY, level.Type_Stairs)

	// //
	// // End cap room 2
	// //
	// eX = x + 1
	// eY = y + 1
	// eWidth = utility.GetRandom(tWidth+2, tWidth+10)
	// eHeight = utility.GetRandom(tHeight+2, tHeight+10)

	// dX = eX - 1
	// dY = eY - 1
	// if wide {
	// 	//eWidth = width - tWidth - 1
	// 	eY = y - tHeight/4
	// 	eWidth = utility.GetRandom(width/8, width/4)
	// 	dY += tHeight / 2
	// 	eX -= eWidth
	// } else {
	// 	//eHeight = height - tHeight - 1
	// 	eX = x - tWidth/4
	// 	eHeight = utility.GetRandom(height/8, height/4)
	// 	dX += tWidth / 2
	// 	eY -= eHeight
	// }

	// CarveRoom(l, eX, eY, eWidth, eHeight, level.Type_Wall, level.Type_Floor, true, true)
	// l.SetTileType(dX, dY, level.Type_Door)

	// Optional second hallway
	if utility.GetRandom(0, 10) > 5 {
		x2 := width / 2
		y2 := height / 2
		tWidth2 := utility.GetRandom(5, 10)
		tHeight2 := utility.GetRandom(5, 10)
		// Makes a +
		if !wide {
			tWidth2 = utility.GetRandom(width/2, width-width/4)
		} else {
			tHeight2 = utility.GetRandom(height/2, height-height/4)
		}

		x2 -= tWidth2 / 2
		y2 -= tHeight2 / 2
		CarveRoom(l, x2, y2, tWidth2, tHeight2, level.Type_Wall, level.Type_Floor, true, false)
		CarveRoom(l, x+1, y+1, tWidth-2, tHeight-2, level.Type_Floor, level.Type_Floor, false, false)     // Erase center of hallway
		CarveRoom(l, x2+1, y2+1, tWidth2-2, tHeight2-2, level.Type_Floor, level.Type_Floor, false, false) // Erase center of hallway

	}

	//Bud rooms
	BudRooms(l, width, height, 50)

	// Polish so we can pathfind
	l.Polish()

	//
	// Maintenance Tunnels
	//
	CarveMaintenanceTunnels(l, width, height, 10)

	l.Polish()
}

func CarveMaintenanceTunnels(l *level.Level, width, height, numTunnels int) {
	for i := 0; i < numTunnels; i++ {
		//fmt.Println("Generating tunnel ", i+1, " out of ", numTunnels)
		done := false
		tries := 0
		for tries < 99999 && !done {
			tries++
			tX1 := utility.GetRandom(0, width)
			tY1 := utility.GetRandom(0, height)

			if l.GetTileType(tX1, tY1) == level.Type_Wall {
				if !l.TileNeighbors(tX1, tY1, level.Type_Open) || !l.TileNeighbors(tX1, tY1, level.Type_Floor) {
					continue
				}
			} else {
				continue
			}

			tX2 := utility.GetRandom(0, width)
			tY2 := utility.GetRandom(0, height)

			if tX2 == tX1 && tY2 == tY1 {
				continue
			}

			if l.GetTileType(tX2, tY2) == level.Type_Wall {
				if !l.TileNeighbors(tX2, tY2, level.Type_Open) || !l.TileNeighbors(tX2, tY2, level.Type_Floor) {
					continue
				}
			} else {
				continue
			}
			l.GetTileAt(tX1, tY1).Solid = false
			l.GetTileAt(tX2, tY2).Solid = false
			steps, distance, success := astar.Path(l.GetTileAt(tX1, tY1), l.GetTileAt(tX2, tY2))

			if !success || distance > 10 || distance < 2 {
				continue
			}

			for step := range steps {
				t := steps[step].(*level.Tile)
				l.SetTileType(t.X, t.Y, level.Type_MaintenanceTunnelFLoor)
			}

			l.SetTileType(tX1, tY1, level.Type_MaintenanceTunnelDoor)
			l.SetTileType(tX2, tY2, level.Type_MaintenanceTunnelDoor)
			done = true
		}
	}
}

func CarveMaintenanceTunnel(l *level.Level, x1, y1, x2, y2 int, floor, door level.TileType) bool {
	l.GetTileAt(x1, y1).Solid = false
	l.GetTileAt(x2, y2).Solid = false
	steps, _, success := astar.Path(l.GetTileAt(x1, y1), l.GetTileAt(x2, y2))

	if !success {
		return false
	}

	for step := range steps {
		t := steps[step].(*level.Tile)
		l.SetTileType(t.X, t.Y, floor)
	}

	l.SetTileType(x1, y1, door)
	l.SetTileType(x2, y2, door)

	return true
}

func BudRooms(l *level.Level, width, height, numRooms int) {
	// Bud rooms
	//numRooms := 50
	for i := 0; i < numRooms; i++ {
		//fmt.Println("Generating room ", i+1, " out of ", numRooms)
		done := false
		tries := 0
		for tries < 99999 && !done {
			tries++
			rX := utility.GetRandom(0, width)
			rY := utility.GetRandom(0, height)

			rHeight := utility.GetRandom(4, 10)
			rWidth := utility.GetRandom(4, 10)

			if rX > 0 && rX < l.Width && rY > 0 && rY < l.Height {
				t := l.GetTileAt(rX, rY)
				if t.NoBudding {
					continue
				}
				if t.Type == level.Type_Wall {
					//find out which direction to build the roomWidth
					//Up
					if l.GetTileType(rX, rY-2) == level.Type_Open {
						if l.GetTileType(rX+1, rY) == level.Type_Wall && l.GetTileType(rX-1, rY) == level.Type_Wall {
							if !RoomIntersects(l, rX-rWidth/2, rY-rHeight-1, rWidth, rHeight) {
								CarveRoom(l, rX-rWidth/2, rY-rHeight+1, rWidth, rHeight, level.Type_Wall, level.Type_Floor, true, false)
								l.SetTileType(rX, rY, level.Type_Door)
								done = true
							}
						}
					}

					//Down
					if l.GetTileType(rX, rY+2) == level.Type_Open {
						if l.GetTileType(rX+1, rY) == level.Type_Wall && l.GetTileType(rX-1, rY) == level.Type_Wall {
							if !RoomIntersects(l, rX-rWidth/2, rY+1, rWidth, rHeight) {
								CarveRoom(l, rX-rWidth/2, rY, rWidth, rHeight, level.Type_Wall, level.Type_Floor, true, false)
								l.SetTileType(rX, rY, level.Type_Door)
								done = true
							}
						}
					}

					// //left
					if l.GetTileType(rX-2, rY) == level.Type_Open {
						if l.GetTileType(rX, rY+1) == level.Type_Wall && l.GetTileType(rX, rY-1) == level.Type_Wall {
							if !RoomIntersects(l, rX-rWidth, rY-rHeight/2, rWidth, rHeight) {
								CarveRoom(l, rX-rWidth+1, rY-rHeight/2, rWidth, rHeight, level.Type_Wall, level.Type_Floor, true, false)
								l.SetTileType(rX, rY, level.Type_Door)
								done = true
							}
						}
					}

					// //right
					if l.GetTileType(rX+2, rY) == level.Type_Open {
						if l.GetTileType(rX, rY+1) == level.Type_Wall && l.GetTileType(rX, rY-1) == level.Type_Wall {
							if !RoomIntersects(l, rX+1, rY-rHeight/2, rWidth, rHeight) {
								CarveRoom(l, rX, rY-rHeight/2, rWidth, rHeight, level.Type_Wall, level.Type_Floor, true, false)
								l.SetTileType(rX, rY, level.Type_Door)
								done = true
							}
						}
					}

				}
			}

		}
	}
}

// Attempts to bud a room off the line provided.   Returns if successful.
func BudRoom(l *level.Level, x1, y1, x2, y2, width, height int) bool {

	return true
}

func RoomIntersects(l *level.Level, x, y, width, height int) bool {
	if x+width > l.Width || y+height > l.Height {
		return true
	}

	if y < 0 || x < 0 {
		return true
	}

	for currentY := 0; currentY < height; currentY++ {
		for currentX := 0; currentX < width; currentX++ {
			t := l.GetTileAt(currentX+x, currentY+y)
			if t == nil {
				return true
			}

			if t.Type != level.Type_Open {
				return true
			}
		}
	}

	return false
}
