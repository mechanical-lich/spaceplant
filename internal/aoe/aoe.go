// Package aoe provides tile-offset generators for common area-of-effect shapes.
// All functions return offsets relative to an origin — callers are responsible
// for adding the origin position and applying whatever effect they need.
package aoe

// Offset is a relative (X, Y) tile displacement from an origin.
type Offset struct{ X, Y int }

// Cone returns tile offsets for a cone in direction (fdx, fdy).
// spread is the perpendicular half-width applied at every depth row.
// Pass spread = -1 for the classic widening behaviour (spread grows by 1 per row):
//
//	depth 1, spread  1 → [1][1][1]            (3 tiles, constant width)
//	depth 2, spread  1 → [1][1][1] / [2][2][2] (3 tiles per row)
//	depth 3, spread -1 → row 1: 1, row 2: 3, row 3: 5 (widening)
func Cone(fdx, fdy, depth, spread int) []Offset {
	pdx, pdy := -fdy, fdx // 90° rotation of the forward vector
	var offsets []Offset
	for d := 1; d <= depth; d++ {
		s := spread
		if s < 0 {
			s = d - 1 // classic widening: grows by 1 each row
		}
		for p := -s; p <= s; p++ {
			offsets = append(offsets, Offset{
				X: fdx*d + pdx*p,
				Y: fdy*d + pdy*p,
			})
		}
	}
	return offsets
}

// Circle returns all tile offsets within the given Euclidean radius,
// excluding the origin itself.
func Circle(radius int) []Offset {
	r2 := radius * radius
	var offsets []Offset
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x == 0 && y == 0 {
				continue
			}
			if x*x+y*y <= r2 {
				offsets = append(offsets, Offset{x, y})
			}
		}
	}
	return offsets
}

// Ring returns tile offsets that fall on the outer shell of the given Euclidean
// radius — a hollow circle. Useful for auras, shockwaves, or area denial.
func Ring(radius int) []Offset {
	outer := radius * radius
	inner := (radius - 1) * (radius - 1)
	var offsets []Offset
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x == 0 && y == 0 {
				continue
			}
			d := x*x + y*y
			if d > inner && d <= outer {
				offsets = append(offsets, Offset{x, y})
			}
		}
	}
	return offsets
}

// Line returns tile offsets in a straight line of the given length in
// direction (fdx, fdy), excluding the origin.
func Line(fdx, fdy, length int) []Offset {
	offsets := make([]Offset, length)
	for i := range offsets {
		offsets[i] = Offset{fdx * (i + 1), fdy * (i + 1)}
	}
	return offsets
}

// Burst returns all 8 adjacent tile offsets (Moore neighbourhood).
// Equivalent to Circle(1) but without the sqrt check and in a fixed order.
func Burst() []Offset {
	return []Offset{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}
}

// DirToVec converts a DirectionComponent.Direction value to a (dx, dy) unit
// vector using the rlcomponents convention: 0=right, 1=down, 2=up, 3=left.
func DirToVec(dir int) (int, int) {
	switch dir {
	case 0:
		return 1, 0
	case 1:
		return 0, 1
	case 2:
		return 0, -1
	default: // 3 = left
		return -1, 0
	}
}
