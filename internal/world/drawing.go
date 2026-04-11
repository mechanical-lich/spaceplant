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
	cfg := config.Global()
	sw := float64(cfg.SpriteSizeW)
	sh := float64(cfg.SpriteSizeH)

	output := ebiten.NewImage(width*cfg.SpriteSizeW, height*cfg.SpriteSizeH)
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
				if cfg.Los {
					seen = rlfov.Los(l.Level, aX, aY, x, y, z)
				}

				l.DrawTile(output, tile, screenX, screenY, seen, z, cfg.SpriteSizeW, cfg.SpriteSizeH, cfg.ColorShading)

				if seen {
					// Draw entities on this tile
					entBuf = entBuf[:0]
					l.Level.GetEntitiesAt(x, y, z, &entBuf)
					tX := float64(screenX) * sw
					tY := float64(screenY) * sh
					for _, entity := range entBuf {
						DrawEntity(output, entity, tX, tY, x, y, cfg.SpriteSizeW, cfg.SpriteSizeH, cfg.ColorShading)
					}

					// Draw tile animations (above entities, below fog)
					l.drawAndAdvanceTileAnims(output, x, y, z, tX, tY, cfg.SpriteSizeW, cfg.SpriteSizeH)

					// Draw fog
					if cfg.Lighting {
						fogColor := color.RGBA{0, 0, 0, uint8(tile.LightLevel)}
						ebitenutil.DrawRect(output, tX, tY, sw, sh, fogColor)
					}
				}
			} else {
				// Draw nothingness if out of bound.
				tX := float64(screenX) * sw
				tY := float64(screenY) * sh
				ebitenutil.DrawRect(output, tX, tY, sw, sh, l.Theme.BackgroundColor)
			}

			screenY++
		}
		screenX++
	}

	// Draw entity paths (debug overlay)
	if cfg.RenderPathfindingSteps {
		dotW := sw / 4
		dotH := sh / 4
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
				sx := float64(tx-left)*sw + sw/2 - dotW/2
				sy := float64(ty-up)*sh + sh/2 - dotH/2
				ebitenutil.DrawRect(output, sx, sy, dotW, dotH, pathColor)
			}
		}
	}

	return output
}

// DrawTile draws a single tile to the output image.
func (l *Level) DrawTile(output *ebiten.Image, t *Tile, screenX, screenY int, seen bool, z, spW, spH int, colorShading bool) {
	sw := float64(spW)
	sh := float64(spH)
	tX := float64(screenX) * sw
	tY := float64(screenY) * sh
	fgColor := l.Theme.TileForgroundColor(t.Type)
	bgColor := l.Theme.TileBackgroundColor(t.Type)

	// Draw background
	ebitenutil.DrawRect(output, tX, tY, sw, sh, bgColor)

	// Draw foreground sprite
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tX, tY)
	if colorShading {
		op.ColorM.ScaleWithColor(fgColor)
	}

	// Resolve sprite coordinates from tile definition
	variant := l.Level.ResolveVariant(t)
	if t.Overgrown {
		// The overgrown sheet has two sub-variants per original tile column.
		// Pick sub-variant deterministically from the tile index so it's stable across frames.
		subVariant := t.Idx % 2
		sX := (variant.SpriteX*2 + subVariant) * spW
		output.DrawImage(resource.Textures["map-overgrown"].SubImage(image.Rect(sX, variant.SpriteY, sX+spW, variant.SpriteY+spH)).(*ebiten.Image), op)
	} else {
		sX := variant.SpriteX * spW
		output.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, variant.SpriteY, sX+spW, variant.SpriteY+spH)).(*ebiten.Image), op)
	}

	tx, ty, _ := t.Coords()
	tileSeen := l.GetSeen(tx, ty, z)

	if !seen {
		if tileSeen {
			ebitenutil.DrawRect(output, tX, tY, sw, sh, color.RGBA{0, 0, 0, 220})
		} else {
			ebitenutil.DrawRect(output, tX, tY, sw, sh, l.Theme.OpenBackgroundColor)
		}
	} else {
		l.SetSeen(tx, ty, z, true)
	}
}

// drawLayered composites the layered sprite system for entities with
// LayeredAppearanceComponent. Layers are drawn in order: body, shirt, pants,
// shoes, hair, headwear. Clothing layers come from equipped items that carry
// WearableAppearanceComponent. Each layer sheet column is spW wide and spH tall.
func drawLayered(screen *ebiten.Image, entity *ecs.Entity, screenX, screenY float64, spW, spH int) {
	lac := entity.GetComponent(component.LayeredAppearance).(*component.LayeredAppearanceComponent)
	bt := lac.BodyType

	dead := entity.HasComponent("Dead")
	drawLayer := func(texName string, index int) {
		tex, ok := resource.Textures[texName]
		if !ok {
			return
		}
		srcX := index * spW
		srcRect := image.Rect(srcX, 0, srcX+spW, spH)
		op := &ebiten.DrawImageOptions{}
		if dead {
			op.GeoM.Scale(1, -1)
			op.GeoM.Translate(0, float64(spH))
		}
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(tex.SubImage(srcRect).(*ebiten.Image), op)
	}

	// Body (skin)
	drawLayer(bt+"_body", lac.BodyIndex)

	// Collect wearable layers from equipped items.
	wearables := map[string]int{}
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.WearableAppearance) {
				wac := item.GetComponent(component.WearableAppearance).(*component.WearableAppearanceComponent)
				wearables[wac.Layer] = wac.Index
			}
		}
	}

	// Draw in fixed layer order: shirt, pants, shoes, then hair, then headwear.
	for _, layer := range []string{"shirt", "pants", "shoes"} {
		if idx, ok := wearables[layer]; ok {
			drawLayer(bt+"_"+layer, idx)
		}
	}
	if lac.HairIndex >= 0 {
		drawLayer(bt+"_hair", lac.HairIndex)
	}
	if idx, ok := wearables["headwear"]; ok {
		drawLayer(bt+"_headwear", idx)
	}
}

// DrawEntity draws a single entity at the given screen position.
// tileWorldX/Y is the world coordinate of the tile being drawn — used to
// select the correct sub-rect of sized entity sprites.
func DrawEntity(screen *ebiten.Image, entity *ecs.Entity, screenX, screenY float64, tileWorldX, tileWorldY, spW, spH int, colorShading bool) {
	if entity == nil {
		return
	}

	// Layered sprite system takes priority over AppearanceComponent.
	if entity.HasComponent(component.LayeredAppearance) {
		drawLayered(screen, entity, screenX, screenY, spW, spH)
		// Still draw FX overlay if present.
		if entity.HasComponent("AttackComponent") {
			attackC := entity.GetComponent("AttackComponent").(*component.AttackComponent)
			if tileWorldX == attackC.X && tileWorldY == attackC.Y {
				if attackC.Frame == 3 {
					entity.RemoveComponent("AttackComponent")
				} else {
					xOffset := attackC.SpriteX + (attackC.Frame * spW)
					fxOp := &ebiten.DrawImageOptions{}
					fxOp.GeoM.Translate(screenX, screenY)
					screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+spW, attackC.SpriteY+spH)).(*ebiten.Image), fxOp)
					attackC.Frame++
				}
			}
		}
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
	sw, sh := spW, spH
	frameX := ac.SpriteX + (entityW * sw * ac.CurrentFrame)
	srcX := frameX + subX*sw
	srcY := ac.SpriteY + subY*sh
	srcRect := image.Rect(srcX, srcY, srcX+sw, srcY+sh)

	op := &ebiten.DrawImageOptions{}

	// Dead transformation — flip the Y sub-tile index so the whole sprite
	// appears upside-down as a unit rather than each tile flipping in place.
	if entity.HasComponent("Dead") {
		subY = (entityH - 1) - subY
		srcY = ac.SpriteY + subY*sh
		srcRect = image.Rect(srcX, srcY, srcX+sw, srcY+sh)
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, float64(sh))
	}

	// Position
	op.GeoM.Translate(screenX, screenY)
	// Color
	if colorShading {
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
				xOffset := attackC.SpriteX + (attackC.Frame * sw)
				fxOp := &ebiten.DrawImageOptions{}
				fxOp.GeoM.Translate(screenX, screenY)
				screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+sw, attackC.SpriteY+sh)).(*ebiten.Image), fxOp)
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
