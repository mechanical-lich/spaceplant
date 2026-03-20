package game

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/ui"
)

// Main gui
type GUIViewMain struct {
	ui.GUIViewBase
	minimap *ebiten.Image
	x       int
}

func (g *GUIViewMain) Update(s any) {
	g.x++
	cs, ok := s.(*SPClientState)
	if ok {
		g.minimap = cs.GetMinimap(0, 0, 100, 100, 150, 150)
	}
}

func (g *GUIViewMain) Draw(screen *ebiten.Image, s any) {
	// Draw Minimap
	if g.minimap != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(config.GameWidth+5, 16)
		screen.DrawImage(g.minimap, op)
	}

	cs, _ := s.(*SPClientState)
	if cs == nil {
		return
	}

	if cs.sim.Player != nil {
		if cs.sim.Player.HasComponent("Health") {
			gc := cs.sim.Player.GetComponent("Health").(*component.HealthComponent)
			mlge_text.Draw(screen, "Hp:"+strconv.Itoa(gc.Health), 24, config.GameWidth, 85+100, color.White)
		}
	}

	for i := 0; i < 10; i++ {
		if i < len(message.MessageLog) {
			m := message.MessageLog[len(message.MessageLog)-1-i]
			mlge_text.Draw(screen, m, 16, config.GameWidth, 85+120+i*32, color.White)
		}
	}

	if config.ShowMouseCoords {
		cX, cY := ebiten.CursorPosition()
		mlge_text.Draw(screen, strconv.Itoa(cX)+","+strconv.Itoa(cY), 16, cX, cY, color.RGBA{255, 0, 0, 255})
	}
}

// GetMinimap generates a minimap image of specified size.
func (cs *SPClientState) GetMinimap(sX, sY, width, height, imageWidth, imageHeight int) *ebiten.Image {
	worldImage := ebiten.NewImage(imageWidth, imageHeight)
	pc := cs.sim.Player.GetComponent("Position").(*component.PositionComponent)

	view := cs.sim.Level.GetView(sX, sY, cs.sim.CurrentZ, width, height, false, false)
	for x := 0; x < len(view); x++ {
		for y := 0; y < len(view[x]); y++ {
			tX := float64(x * imageWidth / width)
			tY := float64(y * imageHeight / height)
			tile := view[x][y]

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tX, tY)

			if tile == nil {
				sX := 19 * config.SpriteWidth
				worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)
				continue
			}
			variant := cs.sim.Level.Level.ResolveVariant(tile)
			spriteX := variant.SpriteX * config.SpriteWidth
			tx, ty, tz := tile.Coords()
			if !cs.sim.Level.GetSeen(tx, ty, tz) {
				spriteX = 19 * config.SpriteWidth
			}
			worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(spriteX, variant.SpriteY, spriteX+config.SpriteWidth, variant.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)
		}
	}

	ebitenutil.DrawRect(worldImage, float64(pc.GetX()*imageWidth/width), float64(pc.GetY()*imageHeight/height), 5, 5, color.RGBA{0, 0, 255, 255})

	return worldImage
}
