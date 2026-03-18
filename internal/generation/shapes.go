package generation

import (
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type Room struct {
	X, Y, Width, Height int
}

func CarveRoom(m *world.Level, x, y, z, width, height, wallType, floorType int, noOverwrite bool, noBudding bool) {
	if x+width > m.Width || y+height > m.Height {
		return
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			t := m.Level.GetTilePtr(x+i, y+j, z)
			if t != nil {
				if noOverwrite {
					if t.Type != world.TypeOpen && t.Type != world.TypeFloor {
						continue
					}
				}
				m.SetNoBudding(x+i, y+j, z, noBudding)
			}

			if i == 0 || i == width-1 || j == 0 || j == height-1 {
				m.SetTileTypeAt(x+i, y+j, z, wallType)
			} else {
				m.SetTileTypeAt(x+i, y+j, z, floorType)
			}
		}
	}
}

func CarveRect(m *world.Level, x1, y1, x2, y2, z, wallType, floorType int, noOverwrite bool, noBudding bool) {
	if x2 > m.Width || y2 > m.Height {
		return
	}

	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			t := m.Level.GetTilePtr(x, y, z)
			if t != nil {
				if noOverwrite {
					if t.Type != world.TypeOpen && t.Type != world.TypeFloor {
						continue
					}
				}
				m.SetNoBudding(x, y, z, noBudding)
			}

			if x == x1 || x == x2 || y == y1 || y == y2 {
				m.SetTileTypeAt(x, y, z, wallType)
			} else {
				m.SetTileTypeAt(x, y, z, floorType)
			}
		}
	}
}

func CarveCircle(m *world.Level, startX, startY, z, r, wallType, floorType int, noOverwrite bool, noBudding bool) {
	for x1 := startX - r; x1 < startX+r; x1++ {
		for y1 := startY - r; y1 < startY+r; y1++ {
			if ((x1-startX)*(x1-startX) + (y1-startY)*(y1-startY)) < r*r {
				if noOverwrite {
					if m.GetTileType(x1, y1, z) != world.TypeOpen {
						continue
					}
				}
				m.SetNoBudding(x1, y1, z, noBudding)
				m.SetTileTypeAt(x1, y1, z, floorType)
				if utility.Distance(x1, y1, startX, startY) >= r-1 {
					m.SetTileTypeAt(x1, y1, z, wallType)
				}
			}
		}
	}
}
