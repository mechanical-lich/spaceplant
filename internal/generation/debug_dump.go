package generation

import (
	"fmt"
	"os"
	"strings"

	"github.com/mechanical-lich/spaceplant/internal/world"
)

// tileChar returns a single ASCII character representing the tile type.
func tileChar(t int) byte {
	switch t {
	case world.TypeFloor:
		return '.'
	case world.TypeWall:
		return '#'
	case world.TypeOpen:
		return ' '
	case world.TypeMaintenanceTunnelFloor:
		return '+'
	case world.TypeMaintenanceTunnelWall:
		return '|'
	case world.TypeMaintenanceTunnelDoor:
		return 'm'
	case world.TypeDoor:
		return 'D'
	case world.TypeStairsUp:
		return '<'
	case world.TypeStairsDown:
		return '>'
	default:
		return '?'
	}
}

// DumpGenerationASCII writes an ASCII map of every floor to gen_debug.txt
// in the current working directory, overwriting any previous run.
func DumpGenerationASCII(l *world.Level, results []FloorResult) {
	var sb strings.Builder

	for _, fr := range results {
		z := fr.Z
		fmt.Fprintf(&sb, "=== Floor %d: %s ===\n", z, fr.Theme.Name)

		for y := 0; y < l.Height; y++ {
			for x := 0; x < l.Width; x++ {
				t := l.Level.GetTilePtr(x, y, z)
				if t == nil {
					sb.WriteByte('?')
				} else {
					sb.WriteByte(tileChar(t.Type))
				}
			}
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')
	}

	if err := os.WriteFile("gen_debug.txt", []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "gen_debug: failed to write: %v\n", err)
	}
}
