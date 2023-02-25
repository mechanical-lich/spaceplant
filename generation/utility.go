package generation

import (
	"github.com/beefsack/go-astar"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/utility"
)

type Room struct {
	X, Y, Width, Height int
}

func CarveRoom(m *level.Level, x, y, width, height int, wallType, floorType level.TileType, noOverwrite bool, noBudding bool) {
	if x+width > m.Width || y+height > m.Height {
		return
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			tile := m.GetTileAt(x+i, y+j)
			if tile != nil {
				if noOverwrite {
					if tile.Type != level.Type_Open && tile.Type != level.Type_Floor {
						continue
					}
				}
				tile.NoBudding = noBudding
			}

			if i == 0 || i == width-1 || j == 0 || j == height-1 {
				m.SetTileType(x+i, y+j, wallType)
			} else {
				m.SetTileType(x+i, y+j, floorType)
			}

		}
	}
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

	//
	// End cap room 1
	//
	eX := x + tWidth - 1
	eY := y + tHeight - 1
	eWidth := utility.GetRandom(tWidth+2, tWidth+10)
	eHeight := utility.GetRandom(tHeight+2, tHeight+10)

	dX := eX
	dY := eY
	if wide {
		//eWidth = width - tWidth - 1
		eY = y - tHeight/4
		eWidth = utility.GetRandom(width/8, width/4)
		dY -= tHeight / 2
	} else {
		//eHeight = height - tHeight - 1
		eX = x - tWidth/4
		eHeight = utility.GetRandom(height/8, height/4)
		dX -= tWidth / 2
	}

	CarveRoom(l, eX, eY, eWidth, eHeight, level.Type_Wall, level.Type_Floor, true, true)
	l.SetTileType(dX, dY, level.Type_Door)

	//
	// End cap room 2
	//
	eX = x + 1
	eY = y + 1
	eWidth = utility.GetRandom(tWidth+2, tWidth+10)
	eHeight = utility.GetRandom(tHeight+2, tHeight+10)

	dX = eX - 1
	dY = eY - 1
	if wide {
		//eWidth = width - tWidth - 1
		eY = y - tHeight/4
		eWidth = utility.GetRandom(width/8, width/4)
		dY += tHeight / 2
		eX -= eWidth
	} else {
		//eHeight = height - tHeight - 1
		eX = x - tWidth/4
		eHeight = utility.GetRandom(height/8, height/4)
		dX += tWidth / 2
		eY -= eHeight
	}

	CarveRoom(l, eX, eY, eWidth, eHeight, level.Type_Wall, level.Type_Floor, true, true)
	l.SetTileType(dX, dY, level.Type_Door)

	// Optional second hallway
	if utility.GetRandom(0, 10) > 8 {
		x = width / 2
		y = height / 2
		tWidth = utility.GetRandom(5, 10)
		tHeight = utility.GetRandom(5, 10)
		// Makes a +
		if !wide {
			tWidth = utility.GetRandom(width/2, width-width/4)
		} else {
			tHeight = utility.GetRandom(height/2, height-height/4)
		}

		x -= tWidth / 2
		y -= tHeight / 2
		CarveRoom(l, x+1, y+1, tWidth-2, tHeight-2, level.Type_Floor, level.Type_Floor, false, false) // Erase center of hallway
		CarveRoom(l, x, y, tWidth, tHeight, level.Type_Wall, level.Type_Floor, true, false)
	}

	// Bud rooms
	numRooms := 50
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

	// Polish so we can pathfind
	l.Polish()

	//
	// Maintenance Tunnels
	//
	numTunnels := 10
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

	l.Polish()

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
