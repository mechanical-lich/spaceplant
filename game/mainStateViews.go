package game

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/mlge/state"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/ui"
)

// Main gui
type GUIViewMain struct {
	ui.GUIViewBase
	minimap *ebiten.Image
	x       int
}

func (g *GUIViewMain) Update(s state.StateInterface) {
	g.x++
	mainState, ok := s.(*MainState)
	if ok {
		g.minimap = mainState.GetMinimap(0, 0, 100, 100, 150, 150)
	}

}

func (g *GUIViewMain) Draw(screen *ebiten.Image, s state.StateInterface) {
	//Draw Minimap
	if g.minimap != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(config.GameWidth+5, 16)
		screen.DrawImage(g.minimap, op)
	}

	mainState, _ := s.(*MainState)

	if mainState.Player != nil {
		if mainState.Player.HasComponent("Health") {
			gc := mainState.Player.GetComponent("Health").(*component.HealthComponent)
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

// GetMinimap generates a minimap image of specified size and returns the image.
func (g *MainState) GetMinimap(sX int, sY int, width int, height int, imageWidth int, imageHeight int) *ebiten.Image {
	worldImage := ebiten.NewImage(imageWidth, imageHeight)
	pc := g.Player.GetComponent("Position").(*component.PositionComponent)

	view := g.level.GetView(sX, sY, g.CurrentZ, width, height, false, false)
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
			} else {
				variant := g.level.Level.ResolveVariant(tile)
				spriteX := variant.SpriteX * config.SpriteWidth
				tx, ty, tz := tile.Coords()
				if !g.level.GetSeen(tx, ty, tz) {
					spriteX = 19 * config.SpriteWidth
				}
				worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(spriteX, variant.SpriteY, spriteX+config.SpriteWidth, variant.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)
			}
		}
	}

	ebitenutil.DrawRect(worldImage, float64(pc.GetX()*imageWidth/width), float64(pc.GetY()*imageHeight/height), 5, 5, color.RGBA{0, 0, 255, 255})

	return worldImage
}
