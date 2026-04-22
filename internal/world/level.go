package world

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
)

// Tile is a direct alias for rlworld.Tile.
type Tile = rlworld.Tile

// Level embeds rlworld.Level and adds spaceplant-specific state.
type Level struct {
	*rlworld.Level
	Theme     Theme
	NoBudding []bool                      // parallel to Level.Data — generation flag
	Overgrown []bool                      // parallel to Level.Data — overgrowth state
	TileAnims map[TileAnimKey][]*TileAnim // visual-only tile overlay animations
	ShaderSrc  []byte                      // KAGE source compiled lazily on first render
	shader     *ebiten.Shader              // compiled from ShaderSrc on first use
	shaderDst  *ebiten.Image              // reused shader output — reallocated only when size changes
	// Flags is a general-purpose key-value store for scripts and scenarios to communicate.
	Flags map[string]any
}

// NewLevel creates a single 3D level.
func NewLevel(width, height, depth int, theme Theme) *Level {
	base := rlworld.NewLevel(width, height, depth)
	total := width * height * depth
	l := &Level{
		Level:     base,
		Theme:     theme,
		NoBudding: make([]bool, total),
		Overgrown: make([]bool, total),
		TileAnims: make(map[TileAnimKey][]*TileAnim),
		Flags:     make(map[string]any),
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
		if blocker := l.Level.GetSolidEntityAt(x, y, z); blocker != nil {
			if blocker.HasComponent(rlcomponents.Door) {
				dc := blocker.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
				if dc.Open {
					return 10 // open doors are passable
				}
			}
			return 100
		}
		return 10
	}

	return l
}

// RevealFloor marks every tile on floor z as seen, fully populating the minimap.
func (l *Level) RevealFloor(z int) {
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			l.SetSeen(x, y, z, true)
		}
	}
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

// IsOvergrown returns whether the tile at (x, y, z) is overgrown.
func (l *Level) IsOvergrown(x, y, z int) bool {
	if !l.Level.InBounds(x, y, z) {
		return false
	}
	return l.Overgrown[l.tileIndex(x, y, z)]
}

// SetOvergrown sets the overgrowth state for the tile at (x, y, z).
func (l *Level) SetOvergrown(x, y, z int, val bool) {
	if !l.Level.InBounds(x, y, z) {
		return
	}
	l.Overgrown[l.tileIndex(x, y, z)] = val
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

// AddTileAnim appends an animation overlay to the tile at (x, y, z).
// Set anim.TTL = -1 for an indefinite animation that never expires on its own.
func (l *Level) AddTileAnim(x, y, z int, anim *TileAnim) {
	key := TileAnimKey{x, y, z}
	l.TileAnims[key] = append(l.TileAnims[key], anim)
}

// ClearTileAnims removes all animations from the tile at (x, y, z).
func (l *Level) ClearTileAnims(x, y, z int) {
	delete(l.TileAnims, TileAnimKey{x, y, z})
}

// drawAndAdvanceTileAnims renders all animations for a tile and removes
// expired ones. Called from the render loop; must not be called concurrently.
func (l *Level) drawAndAdvanceTileAnims(dst *ebiten.Image, x, y, z int, sx, sy float64, spW, spH int) {
	key := TileAnimKey{x, y, z}
	anims := l.TileAnims[key]
	if len(anims) == 0 {
		return
	}
	live := anims[:0]
	for _, a := range anims {
		a.draw(dst, sx, sy, spW, spH)
		if a.LightLevel > 0 {
			l.GetTilePtr(x, y, z).LightLevel = 255 - a.LightLevel
		}
		if !a.advance() {
			live = append(live, a)
		}
	}
	if len(live) == 0 {
		delete(l.TileAnims, key)
	} else {
		l.TileAnims[key] = live
	}
}
