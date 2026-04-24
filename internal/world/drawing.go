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
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// Render renders the level viewport centered on (aX, aY) at Z-layer z.
func (l *Level) Render(aX, aY, z, width, height int, blind, centered bool) *ebiten.Image {
	cfg := config.Global()
	sw := float64(cfg.TileSizeW)
	sh := float64(cfg.TileSizeH)

	output := ebiten.NewImage(width*cfg.TileSizeW, height*cfg.TileSizeH)
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

				l.DrawTile(output, tile, screenX, screenY, seen, z, cfg.TileSizeW, cfg.TileSizeH, cfg.ColorShading)

				if seen {
					// Draw entities on this tile
					entBuf = entBuf[:0]
					l.Level.GetEntitiesAt(x, y, z, &entBuf)
					tX := float64(screenX) * sw
					tY := float64(screenY) * sh
					for _, entity := range entBuf {
						DrawEntity(output, entity, tX, tY, x, y, cfg.TileSizeW, cfg.TileSizeH, cfg.ColorShading)
					}

					// Draw tile animations (above entities, below fog)
					l.drawAndAdvanceTileAnims(output, x, y, z, tX, tY, cfg.TileSizeW, cfg.TileSizeH)

					// Draw fog, covering the full sprite area which may exceed the tile grid cell.
					if cfg.Lighting {
						fogColor := color.RGBA{0, 0, 0, uint8(tile.LightLevel)}
						def := rlworld.TileDefinitions[tile.Type]
						fogY := tY + float64(def.SpriteOffsetY)
						fogH := sh
						if def.SpriteHeight > cfg.TileSizeH {
							fogH = float64(def.SpriteHeight)
						}
						ebitenutil.DrawRect(output, tX, fogY, sw, fogH, fogColor)
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


	if l.shader == nil && l.ShaderSrc != nil {
		if s, err := ebiten.NewShader(l.ShaderSrc); err == nil {
			l.shader = s
		}
	}
	if l.shader != nil {
		w, h := output.Bounds().Dx(), output.Bounds().Dy()
		if l.shaderDst == nil || l.shaderDst.Bounds().Dx() != w || l.shaderDst.Bounds().Dy() != h {
			l.shaderDst = ebiten.NewImage(w, h)
		}
		l.shaderDst.Clear()
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = output
		op.Uniforms = map[string]any{"Intensity": float32(config.Global().CRTIntensity)}
		l.shaderDst.DrawRectShader(w, h, l.shader, op)
		return l.shaderDst
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
	tx, ty, _ := t.Coords()
	def := rlworld.TileDefinitions[t.Type]
	texName := def.Resource
	if texName == "" {
		texName = "map"
	}

	// Use definition-level sprite dimensions/offset if specified; fall back to tile size.
	sprW, sprH := spW, spH
	if def.SpriteWidth > 0 {
		sprW = def.SpriteWidth
	}
	if def.SpriteHeight > 0 {
		sprH = def.SpriteHeight
	}
	if def.SpriteOffsetX != 0 || def.SpriteOffsetY != 0 {
		op.GeoM.Reset()
		op.GeoM.Translate(tX+float64(def.SpriteOffsetX), tY+float64(def.SpriteOffsetY))
	}

	if l.IsOvergrown(tx, ty, z) {
		// The overgrown sheet has two sub-variants per original tile column.
		// Pick sub-variant deterministically from the tile index so it's stable across frames.
		subVariant := t.Idx % 2
		sX := (variant.SpriteX*2 + subVariant) * spW
		output.DrawImage(resource.Textures[texName+"-overgrown"].SubImage(image.Rect(sX, variant.SpriteY, sX+sprW, variant.SpriteY+sprH)).(*ebiten.Image), op)
	} else {
		sX := variant.SpriteX * spW
		output.DrawImage(resource.Textures[texName].SubImage(image.Rect(sX, variant.SpriteY, sX+sprW, variant.SpriteY+sprH)).(*ebiten.Image), op)
	}

	tileSeen := l.GetSeen(tx, ty, z)

	// Cover the full sprite area, which may be taller than the tile grid cell.
	overlayY := tY + float64(def.SpriteOffsetY)
	overlayH := sh
	if sprH > spH {
		overlayH = float64(sprH)
		overlayY = tY + float64(def.SpriteOffsetY)
	}

	if !seen {
		if tileSeen {
			ebitenutil.DrawRect(output, tX, overlayY, sw, overlayH, color.RGBA{0, 0, 0, 220})
		} else {
			ebitenutil.DrawRect(output, tX, overlayY, sw, overlayH, l.Theme.OpenBackgroundColor)
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

	// Layered sprites are 32x48 drawn on 32x32 tiles — shift up 16px so the
	// sprite's feet land on the tile rather than overflowing below it.
	const spriteH = 48
	offsetY := float64(spH - spriteH)

	dead := entity.HasComponent("Dead")
	drawLayer := func(texName string, index int) {
		tex, ok := resource.Textures[texName]
		if !ok {
			return
		}
		srcX := index * spW
		srcRect := image.Rect(srcX, 0, srcX+spW, spriteH)
		op := &ebiten.DrawImageOptions{}
		if dead {
			op.GeoM.Scale(1, -1)
			op.GeoM.Translate(0, float64(spriteH))
		}
		op.GeoM.Translate(screenX, screenY+offsetY)
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
		if entity.HasComponent(rlcomponents.AsciiAppearance) {
			ac := entity.GetComponent(rlcomponents.AsciiAppearance).(*rlcomponents.AsciiAppearanceComponent)
			fontSize := float64(min(spW, spH)) * 0.75
			w, h := mlge_text.Measure(ac.Character, fontSize)
			x := screenX + (float64(spW)-w)/2
			y := screenY + (float64(spH)-h)/2
			mlge_text.Draw(screen, ac.Character, fontSize, int(x), int(y), color.RGBA{ac.R, ac.G, ac.B, 255})
		}
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

	// Per-tile slice of the sprite. For single entities this equals SpriteWidth/Height.
	// For multi-tile entities (e.g. 2×2 with a 64×96 sprite) each tile draws its
	// own SpriteWidth/entityW × SpriteHeight/entityH portion.
	tileW := ac.SpriteWidth / entityW
	tileH := ac.SpriteHeight / entityH
	frameX := ac.SpriteX + (ac.SpriteWidth * ac.CurrentFrame)
	srcX := frameX + subX*tileW
	srcY := ac.SpriteY + subY*tileH
	srcRect := image.Rect(srcX, srcY, srcX+tileW, srcY+tileH)

	op := &ebiten.DrawImageOptions{}

	// Dead transformation — flip the Y sub-tile index so the whole sprite
	// appears upside-down as a unit rather than each tile flipping in place.
	if entity.HasComponent("Dead") {
		subY = (entityH - 1) - subY
		srcY = ac.SpriteY + subY*tileH
		srcRect = image.Rect(srcX, srcY, srcX+tileW, srcY+tileH)
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, float64(tileH))
	}

	// Position — apply per-sprite pixel offset on top of tile position
	op.GeoM.Translate(screenX+float64(ac.SpriteOffsetX), screenY+float64(ac.SpriteOffsetY))
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
				xOffset := attackC.SpriteX + (attackC.Frame * spW)
				fxOp := &ebiten.DrawImageOptions{}
				fxOp.GeoM.Translate(screenX, screenY)
				screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+spW, attackC.SpriteY+spH)).(*ebiten.Image), fxOp)
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
