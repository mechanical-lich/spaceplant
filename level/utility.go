package level

import (
	"math"

	"github.com/mechanical-lich/spaceplant/utility"
)

func Los(pX int, pY int, tX int, tY int, level *Level) bool {
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
			if level.IsTileSolid(tX, tY) || level.GetTileType(tX, tY) == Type_Door || level.GetTileType(tX, tY) == Type_MaintenanceTunnelDoor {
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

		if level.IsTileSolid(tX, tY) || level.GetTileType(tX, tY) == Type_Door || level.GetTileType(tX, tY) == Type_MaintenanceTunnelDoor {
			break
		}
	}

	return false

}
