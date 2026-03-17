package game

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"

	"github.com/mechanical-lich/game-engine/state"
	text_ext "github.com/mechanical-lich/game-engine/text"
	"github.com/mechanical-lich/game-engine/ui"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/message"
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

	// //Draw buttons
	// for _, b := range g.Buttons {
	// 	b.Draw(screen)
	// }
	mainState, _ := s.(*MainState)

	if mainState.Player != nil {
		if mainState.Player.HasComponent("HealthComponent") {
			gc := mainState.Player.GetComponent("HealthComponent").(*component.HealthComponent)
			text.Draw(screen, "Hp:"+strconv.Itoa(gc.Health), text_ext.MplusNormalFont, config.GameWidth, 85+100, color.White)
		}
	}

	for i := 0; i < 10; i++ {
		if i < len(message.MessageLog) {
			m := message.MessageLog[len(message.MessageLog)-1-i]
			text.Draw(screen, m, text_ext.MplusSmallFont, config.GameWidth, 85+120+i*32, color.White)
		}

	}

	// if mainState.showInventory {
	// 	cX, cY := ebiten.CursorPosition()

	// 	dialogX := 375.0
	// 	dialogY := 225.0
	// 	dialogWidth := 500.0
	// 	dialogHeight := 500.0
	// 	ebitenutil.DrawRect(screen, dialogX, dialogY, dialogWidth, dialogHeight, color.RGBA{50, 50, 50, 200})
	// 	if mainState.Player != nil {
	// 		inventory := mainState.Player.GetComponent("InventoryComponent").(*component.InventoryComponent)
	// 		for i, v := range inventory.Bag {
	// 			d := v.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
	// 			m := fmt.Sprintf("(%d) - %s", i, d.Name)

	// 			if cX > int(dialogX) && cX < int(dialogX+dialogWidth) && cY < int(dialogY)+16+(i*16) && cY > int(dialogY)+(i*16) {
	// 				ebitenutil.DrawRect(screen, dialogX+5.0, dialogY+float64((i*16)), dialogWidth, 16, color.RGBA{0, 50, 50, 200})
	// 				text.Draw(screen, m, text_ext.MplusSmallFont, int(dialogX)+5, int(dialogY)+(i*16)+16, color.Black)

	// 				// TODO Temp use code
	// 				if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
	// 					inventory.Use(v, mainState.Player)
	// 				}
	// 			} else {
	// 				text.Draw(screen, m, text_ext.MplusSmallFont, int(dialogX)+5, int(dialogY)+(i*16)+16, color.White)
	// 			}
	// 		}
	// 	}
	// }

	if config.ShowMouseCoords {
		cX, cY := ebiten.CursorPosition()
		text.Draw(screen, strconv.Itoa(cX)+","+strconv.Itoa(cY), text_ext.MplusSmallFont, cX, cY, color.RGBA{255, 0, 0, 255})
	}

}
