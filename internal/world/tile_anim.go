package world

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/resource"
)

// TileAnimKey is the world-space coordinate key for tile animations.
type TileAnimKey struct{ X, Y, Z int }

// TileAnim is a sprite-sheet animation drawn on top of a tile.
// It is advanced and culled during the render pass (not the simulation pass).
type TileAnim struct {
	// SpriteX/SpriteY is the pixel origin of the first frame on the sheet.
	SpriteX, SpriteY int
	// Resource is the texture name to sample from (defaults to "fx").
	Resource string
	// FrameCount is the total number of animation frames.
	FrameCount int
	// FrameSpeed is the number of render ticks between frame advances.
	// 1 = advance every frame; higher values slow the animation down.
	FrameSpeed int
	// TTL is the remaining render ticks before the animation is removed.
	// Set to -1 for an indefinite animation that never expires.
	TTL int

	// internal state
	frame int
	tick  int
}

// advance steps the animation forward by one render tick.
// Returns true if the animation has expired and should be removed.
func (a *TileAnim) advance() bool {
	if a.TTL == 0 {
		return true
	}
	if a.TTL > 0 {
		a.TTL--
	}
	a.tick++
	if a.FrameSpeed < 1 {
		a.FrameSpeed = 1
	}
	if a.tick >= a.FrameSpeed {
		a.tick = 0
		a.frame = (a.frame + 1) % a.FrameCount
	}
	return false
}

// draw renders the current animation frame at screen position (sx, sy).
func (a *TileAnim) draw(dst *ebiten.Image, sx, sy float64, spW, spH int) {
	tex := a.Resource
	if tex == "" {
		tex = "fx"
	}
	t, ok := resource.Textures[tex]
	if !ok {
		return
	}
	x0 := a.SpriteX + a.frame*spW
	y0 := a.SpriteY
	src := image.Rect(x0, y0, x0+spW, y0+spH)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(sx, sy)
	dst.DrawImage(t.SubImage(src).(*ebiten.Image), op)
}
