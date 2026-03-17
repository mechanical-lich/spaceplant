package world

import (
	"math"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/spaceplant/utility"
)

// Los checks line of sight from (pX, pY) to (tX, tY) on the given Z-layer.
func Los(pX, pY, tX, tY, z int, level *Level) bool {
	deltaX := pX - tX
	deltaY := pY - tY

	absDeltaX := math.Abs(float64(deltaX))
	absDeltaY := math.Abs(float64(deltaY))

	signX := utility.Sgn(deltaX)
	signY := utility.Sgn(deltaY)

	if absDeltaX > absDeltaY {
		t := absDeltaY*2 - absDeltaX
		for {
			if t >= 0 {
				tY += signY
				t -= absDeltaX * 2
			}

			tX += signX
			t += absDeltaY * 2

			if tX == pX && tY == pY {
				return true
			}
			tile := level.Level.GetTilePtr(tX, tY, z)
			if tile == nil {
				break
			}
			if tile.IsSolid() {
				break
			}
			def := rlworld.TileDefinitions[tile.Type]
			if def.Door {
				break
			}
		}
		return false
	}

	t := absDeltaX*2 - absDeltaY

	for {
		if t >= 0 {
			tX += signX
			t -= absDeltaY * 2
		}
		tY += signY
		t += absDeltaX * 2
		if tX == pX && tY == pY {
			return true
		}

		tile := level.Level.GetTilePtr(tX, tY, z)
		if tile == nil {
			break
		}
		if tile.IsSolid() {
			break
		}
		def := rlworld.TileDefinitions[tile.Type]
		if def.Door {
			break
		}
	}

	return false
}
