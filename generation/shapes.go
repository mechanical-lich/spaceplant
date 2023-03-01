package generation

import (
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

func CarveCircle(m *level.Level, start_X, start_Y, r int, wallType, floorType level.TileType, noOverwrite bool, noBudding bool) {
	//width := r * 2
	//height := r * 2

	for x1 := start_X - r; x1 < start_X+r; x1++ {
		for y1 := start_Y - r; y1 < start_Y+r; y1++ {
			if ((x1-start_X)*(x1-start_X) + (y1-start_Y)*(y1-start_Y)) < r*r {
				if noOverwrite {
					if m.GetTileType(x1, y1) != level.Type_Open {
						continue
					}
				}
				m.SetTileNoBudding(x1, y1, noBudding)
				m.SetTileType(x1, y1, floorType)
				if utility.Distance(x1, y1, start_X, start_Y) >= r-1 {
					m.SetTileType(x1, y1, wallType)
				}
			}

		}
	}
}
