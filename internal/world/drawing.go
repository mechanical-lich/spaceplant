package world

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
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
					seen = rlfov.Los(l.Level, aX, aY, x, y, z)
				}

				l.DrawTile(output, tile, screenX, screenY, seen, z)

				if seen {
					// Draw entities on this tile
					entBuf = entBuf[:0]
					l.Level.GetEntitiesAt(x, y, z, &entBuf)
					tX := float64(screenX * config.SpriteWidth)
					tY := float64(screenY * config.SpriteHeight)
					for _, entity := range entBuf {
						DrawEntity(output, entity, tX, tY, x, y)
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

	// Draw entity paths (debug overlay)
	if config.DrawEntityPaths {
		dotW := float64(config.SpriteWidth) / 4
		dotH := float64(config.SpriteHeight) / 4
		pathColor := color.RGBA{255, 220, 0, 200}
		for _, entity := range l.Level.Entities {
			if !entity.HasComponent(component.HostileAI) {
				continue
			}
			pc := entity.GetComponent(component.Position).(*component.PositionComponent)
			if pc.GetZ() != z {
				continue
			}
			hc := entity.GetComponent(component.HostileAI).(*component.HostileAIComponent)
			for _, tileIdx := range hc.Path {
				tile := l.Level.GetTilePtrIndex(tileIdx)
				if tile == nil {
					continue
				}
				tx, ty, tz := tile.Coords()
				if tz != z {
					continue
				}
				sx := float64((tx-left)*config.SpriteWidth) + float64(config.SpriteWidth)/2 - dotW/2
				sy := float64((ty-up)*config.SpriteHeight) + float64(config.SpriteHeight)/2 - dotH/2
				ebitenutil.DrawRect(output, sx, sy, dotW, dotH, pathColor)
			}
		}
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
// tileWorldX/Y is the world coordinate of the tile being drawn — used to
// select the correct sub-rect of sized entity sprites.
func DrawEntity(screen *ebiten.Image, entity *ecs.Entity, screenX, screenY float64, tileWorldX, tileWorldY int) {
	if entity == nil {
		return
	}
	if !entity.HasComponent("AppearanceComponent") {
		return
	}
	ac := entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent)

	// Determine entity size and which sub-tile of the sprite to draw.
	entityW, entityH := 1, 1
	subX, subY := 0, 0
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		if sc.Width > 0 {
			entityW = sc.Width
		}
		if sc.Height > 0 {
			entityH = sc.Height
		}
		pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		startX := pc.GetX() - entityW/2
		startY := pc.GetY() - entityH/2
		subX = tileWorldX - startX
		subY = tileWorldY - startY
	}

	// Compute source rect on the sprite sheet.
	// For sized entities each animation frame spans entityW*SpriteWidth pixels.
	frameX := ac.SpriteX + (entityW * config.SpriteWidth * ac.CurrentFrame)
	srcX := frameX + subX*config.SpriteWidth
	srcY := ac.SpriteY + subY*config.SpriteHeight
	srcRect := image.Rect(srcX, srcY, srcX+config.SpriteWidth, srcY+config.SpriteHeight)

	op := &ebiten.DrawImageOptions{}

	// Dead transformation — flip the Y sub-tile index so the whole sprite
	// appears upside-down as a unit rather than each tile flipping in place.
	if entity.HasComponent("Dead") {
		subY = (entityH - 1) - subY
		srcY = ac.SpriteY + subY*config.SpriteHeight
		srcRect = image.Rect(srcX, srcY, srcX+config.SpriteWidth, srcY+config.SpriteHeight)
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, float64(config.SpriteHeight))
	}

	// Position
	op.GeoM.Translate(screenX, screenY)
	// Color
	if config.ColorShading {
		op.ColorM.ScaleWithColor(color.RGBA{ac.R, ac.G, ac.B, 255})
	}

	texName := ac.Resource
	if texName == "" {
		texName = "entities"
	}
	screen.DrawImage(resource.Textures[texName].SubImage(srcRect).(*ebiten.Image), op)

	// Draw FX on the tile where the hit landed.
	if entity.HasComponent("AttackComponent") {
		attackC := entity.GetComponent("AttackComponent").(*component.AttackComponent)
		if tileWorldX == attackC.X && tileWorldY == attackC.Y {
			if attackC.Frame == 3 {
				entity.RemoveComponent("AttackComponent")
			} else {
				xOffset := attackC.SpriteX + (attackC.Frame * config.SpriteWidth)
				fxOp := &ebiten.DrawImageOptions{}
				fxOp.GeoM.Translate(screenX, screenY)
				screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+config.SpriteWidth, attackC.SpriteY+config.SpriteHeight)).(*ebiten.Image), fxOp)
				attackC.Frame++
			}
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
