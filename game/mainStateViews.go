package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/mechanical-lich/mlge/state"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/message"
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
		//if g.minimap == nil {
		g.minimap = mainState.GetMinimap(0, 0, 100, 100, 150, 150)
		//}
	}

}

func (g *GUIViewMain) Draw(screen *ebiten.Image, s state.StateInterface) {
	//Draw Minimap
	if g.minimap != nil {
		op := &ebiten.DrawImageOptions{}
		//op.GeoM.Scale(.2, .2)
		op.GeoM.Translate(config.GameWidth+5, 16)
		screen.DrawImage(g.minimap, op)
	}

	mainState, _ := s.(*MainState)

	if mainState.Player != nil {
		if mainState.Player.HasComponent("HealthComponent") {
			gc := mainState.Player.GetComponent("HealthComponent").(*component.HealthComponent)
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
