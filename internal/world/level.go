package world

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
)

// Tile is a direct alias for rlworld.Tile.
type Tile = rlworld.Tile

// Level embeds rlworld.Level and adds spaceplant-specific state.
type Level struct {
	*rlworld.Level
	Theme     Theme
	NoBudding []bool // parallel to Level.Data — generation flag
}

// NewLevel creates a single 3D level.
func NewLevel(width, height, depth int, theme Theme) *Level {
	base := rlworld.NewLevel(width, height, depth)
	total := width * height * depth
	l := &Level{
		Level:     base,
		Theme:     theme,
		NoBudding: make([]bool, total),
	}

	// Initialize all tiles to "open" (space)
	for i := range l.Level.Data {
		l.Level.Data[i].Type = TypeOpen
	}

	// Set custom path cost
	base.PathCostFunc = func(from, to *Tile) float64 {
		if to.IsSolid() {
			return 100
		}
		def := rlworld.TileDefinitions[to.Type]
		if def.Air {
			return 0 // open space is free for pathfinding
		}
		x, y, z := to.Coords()
		if l.Level.GetSolidEntityAt(x, y, z) != nil {
			return 100
		}
		return 10
	}

	return l
}

// IsTileSolid checks whether a tile is solid at the given coords.
func (l *Level) IsTileSolid(x, y, z int) bool {
	t := l.Level.GetTilePtr(x, y, z)
	if t == nil {
		return false
	}
	return t.IsSolid()
}

// GetTileType returns the tile type index at (x, y, z), or -1 if out of bounds.
func (l *Level) GetTileType(x, y, z int) int {
	t := l.Level.GetTilePtr(x, y, z)
	if t == nil {
		return -1
	}
	return t.Type
}

// SetTileTypeAt sets the tile type by name at (x, y, z).
func (l *Level) SetTileTypeAt(x, y, z int, tileType int) error {
	t := l.Level.GetTilePtr(x, y, z)
	if t == nil {
		return nil
	}
	t.Type = tileType
	return nil
}

// TileNeighbors checks if any cardinal neighbor has the given type.
func (l *Level) TileNeighbors(x, y, z int, nType int) bool {
	for _, offset := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		if l.GetTileType(x+offset[0], y+offset[1], z) == nType {
			return true
		}
	}
	return false
}

// GetNoBudding returns the NoBudding flag for a tile.
func (l *Level) GetNoBudding(x, y, z int) bool {
	if !l.Level.InBounds(x, y, z) {
		return false
	}
	return l.NoBudding[l.tileIndex(x, y, z)]
}

// SetNoBudding sets the NoBudding flag for a tile.
func (l *Level) SetNoBudding(x, y, z int, val bool) {
	if !l.Level.InBounds(x, y, z) {
		return
	}
	l.NoBudding[l.tileIndex(x, y, z)] = val
}

func (l *Level) tileIndex(x, y, z int) int {
	return x + y*l.Width + z*l.Width*l.Height
}
