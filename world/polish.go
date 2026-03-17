package world

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/spaceplant/utility"
)

// Polish performs post-generation tile variant selection for a single Z-layer.
func (l *Level) Polish(z int) {
	// Build tunnel walls around maintenance tunnel floors
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			t := l.Level.GetTilePtr(x, y, z)
			if t != nil && t.Type == TypeOpen {
				if l.TileNeighbors(x, y, z, TypeMaintenanceTunnelFloor) {
					t.Type = TypeMaintenanceTunnelWall
				}
			}
		}
	}

	// Assign variants
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			t := l.Level.GetTilePtr(x, y, z)
			if t == nil {
				continue
			}
			def := rlworld.TileDefinitions[t.Type]

			switch t.Type {
			case TypeWall:
				// Pick random wall variant (indices 0-9 are normal walls)
				t.Variant = utility.GetRandom(0, 10)
				belowTile := l.Level.GetTilePtr(x, y+1, z)
				if belowTile != nil {
					bt := belowTile.Type
					if bt == TypeWall || bt == TypeDoor || bt == TypeMaintenanceTunnelDoor || bt == TypeMaintenanceTunnelWall {
						// Wall top variant (index 10)
						t.Variant = 10
					}
					if bt == TypeOpen {
						// Wall outside variant (index 11, which maps to spriteX 4)
						t.Variant = 11
					}
				}

			case TypeFloor:
				// Heavily weighted toward variant 0 (spriteX 15)
				if utility.GetRandom(0, 33) < 30 {
					t.Variant = 0
				} else {
					t.Variant = 1
				}

			case TypeDoor:
				t.Variant = utility.GetRandom(0, len(def.Variants))

			case TypeMaintenanceTunnelWall:
				t.Variant = 0
				belowTile := l.Level.GetTilePtr(x, y+1, z)
				if belowTile != nil {
					bt := belowTile.Type
					if bt == TypeMaintenanceTunnelWall || bt == TypeMaintenanceTunnelDoor || bt == TypeDoor {
						t.Variant = 1 // tunnel top
					}
				}

			case TypeMaintenanceTunnelFloor:
				t.Variant = 0

			case TypeStairsUp:
				t.Variant = 0

			case TypeStairsDown:
				t.Variant = 0

			case TypeOpen:
				t.Variant = 0

			case TypeMaintenanceTunnelDoor:
				t.Variant = 0

			default:
				if len(def.Variants) > 0 {
					t.Variant = utility.GetRandom(0, len(def.Variants))
				}
			}
		}
	}
}
