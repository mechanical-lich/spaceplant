package world

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
)

// Render renders the level viewport centered on (aX, aY) at Z-layer z.
func (l *Level) Render(aX, aY, z, width, height int, blind, centered bool) *ebiten.Image {
	output := ebiten.NewImage(width*config.SpriteWidth, height*config.SpriteHeight)
	left := aX - width/2
	right := aX + width/2
	up := aY - height/2
	down := aY + height/2

	if !centered {
		left = aX
		right = aX + width - 1
		up = aY
		down = aY + height
	}

	screenX := 0
	screenY := 0
	var entBuf []*ecs.Entity
	for x := left; x <= right; x++ {
		screenY = 0
		for y := up; y <= down; y++ {
			tile := l.Level.GetTilePtr(x, y, z)

			if blind {
				if y < aY-height/4 || y > aY+height/4 || x > aX+width/4 || x < aX-width/4 {
					tile = nil
				}
			}

			if tile != nil {
				seen := true
				if config.Los {
					seen = Los(aX, aY, x, y, z, l)
				}

				l.DrawTile(output, tile, screenX, screenY, seen, z)

				if seen {
					// Draw entities on this tile
					entBuf = entBuf[:0]
					l.Level.GetEntitiesAt(x, y, z, &entBuf)
					tX := float64(screenX * config.SpriteWidth)
					tY := float64(screenY * config.SpriteHeight)
					for _, entity := range entBuf {
						DrawEntity(output, entity, tX, tY)
					}

					// Draw fog
					if config.Lighting {
						fogColor := color.RGBA{0, 0, 0, uint8(tile.LightLevel)}
						tX := float64(screenX * config.SpriteWidth)
						tY := float64(screenY * config.SpriteHeight)
						ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, fogColor)
					}
				}
			} else {
				// Draw nothingness if out of bound.
				tX := float64(screenX * config.SpriteWidth)
				tY := float64(screenY * config.SpriteHeight)
				ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, l.Theme.BackgroundColor)
			}

			screenY++
		}
		screenX++
	}

	return output
}

// DrawTile draws a single tile to the output image.
func (l *Level) DrawTile(output *ebiten.Image, t *Tile, screenX, screenY int, seen bool, z int) {
	tX := float64(screenX * config.SpriteWidth)
	tY := float64(screenY * config.SpriteHeight)
	fgColor := l.Theme.TileForgroundColor(t.Type)
	bgColor := l.Theme.TileBackgroundColor(t.Type)

	// Draw background
	ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, bgColor)

	// Draw foreground sprite
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tX, tY)
	if config.ColorShading {
		op.ColorM.ScaleWithColor(fgColor)
	}

	// Resolve sprite coordinates from tile definition
	variant := l.Level.ResolveVariant(t)
	sX := variant.SpriteX * config.SpriteWidth
	output.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, variant.SpriteY, sX+config.SpriteWidth, variant.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)

	tx, ty, _ := t.Coords()
	tileSeen := l.GetSeen(tx, ty, z)

	if !seen {
		if tileSeen {
			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, color.RGBA{0, 0, 0, 220})
		} else {
			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, l.Theme.OpenBackgroundColor)
		}
	} else {
		l.SetSeen(tx, ty, z, true)
	}
}

// DrawEntity draws a single entity at the given screen position.
func DrawEntity(screen *ebiten.Image, entity *ecs.Entity, x, y float64) {
	if entity == nil {
		return
	}
	if !entity.HasComponent("AppearanceComponent") {
		return
	}
	ac := entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent)

	op := &ebiten.DrawImageOptions{}

	// Dead transformation
	if entity.HasComponent("Dead") {
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, float64(config.SpriteHeight))
	}

	// Position
	op.GeoM.Translate(x, y)
	// Color
	if config.ColorShading {
		op.ColorM.ScaleWithColor(color.RGBA{ac.R, ac.G, ac.B, 255})
	}

	screen.DrawImage(resource.Textures["entities"].SubImage(image.Rect(ac.GetFrameX(), ac.SpriteY, ac.GetFrameX()+config.SpriteWidth, ac.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)

	// Draw FX
	if entity.HasComponent("AttackComponent") {
		attackC := entity.GetComponent("AttackComponent").(*component.AttackComponent)
		if attackC.Frame == 3 {
			entity.RemoveComponent("AttackComponent")
		} else {
			xOffset := attackC.SpriteX + (attackC.Frame * config.SpriteWidth)
			fxOp := &ebiten.DrawImageOptions{}
			fxOp.GeoM.Translate(x, y)
			screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+config.SpriteWidth, attackC.SpriteY+config.SpriteHeight)).(*ebiten.Image), fxOp)
			attackC.Frame++
		}
	}
}

// GetView returns a 2D slice of tile pointers for the viewport. Used by minimap.
func (l *Level) GetView(aX, aY, z, width, height int, blind, centered bool) [][]*rlworld.Tile {
	left := aX - width/2
	right := aX + width/2
	up := aY - height/2
	down := aY + height/2

	if !centered {
		left = aX
		right = aX + width - 1
		up = aY
		down = aY + height
	}

	data := make([][]*rlworld.Tile, width+1-width%2)

	cX := 0
	for x := left; x <= right; x++ {
		col := []*rlworld.Tile{}
		for y := up; y <= down; y++ {
			currentTile := l.Level.GetTilePtr(x, y, z)
			if blind {
				if y < aY-height/4 || y > aY+height/4 || x > aX+width/4 || x < aX-width/4 {
					currentTile = nil
				}
			}
			col = append(col, currentTile)
		}
		data[cX] = append(data[cX], col...)
		cX++
	}
	return data
}
