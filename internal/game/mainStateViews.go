package game

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/ui"
)

const msgPanelH = 250

// Main gui
type GUIViewMain struct {
	ui.GUIViewBase
	minimap    *ebiten.Image
	msgArea    *minui.ScrollingTextArea
	msgSynced  int // number of MessageLog entries already pushed to msgArea
}

func (g *GUIViewMain) initMsgArea() {
	if g.msgArea != nil {
		return
	}
	cfg := config.Global()
	panelX := cfg.WorldWidth + 4
	panelW := cfg.ScreenWidth - cfg.WorldWidth - 8
	panelY := cfg.ScreenHeight - msgPanelH - 4
	g.msgArea = minui.NewScrollingTextArea("msglog", panelW, msgPanelH)
	g.msgArea.SetPosition(panelX, panelY)
	g.msgArea.LineHeight = 15
}

func (g *GUIViewMain) Update(s any) {
	g.initMsgArea()
	cs, ok := s.(*SPClientState)
	if ok {
		g.minimap = cs.GetMinimap(0, 0, 100, 100, 150, 150)
	}
	// Append any new messages (MessageLog grows monotonically).
	for g.msgSynced < len(message.MessageLog) {
		g.msgArea.AddText(message.MessageLog[g.msgSynced])
		g.msgSynced++
	}
	g.msgArea.Update()
}

func (g *GUIViewMain) Draw(screen *ebiten.Image, s any) {
	// Draw Minimap
	if g.minimap != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(config.Global().WorldWidth)+5, 16)
		screen.DrawImage(g.minimap, op)
	}

	cs, _ := s.(*SPClientState)
	if cs == nil {
		return
	}

	if cs.sim.Player != nil && cs.sim.Player.HasComponent(component.Body) {
		bc := cs.sim.Player.GetComponent(component.Body).(*component.BodyComponent)
		keys := make([]string, 0, len(bc.Parts))
		for k := range bc.Parts {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		y := 185
		for _, name := range keys {
			part := bc.Parts[name]
			var label string
			var col color.RGBA
			if part.Amputated {
				label = fmt.Sprintf("%s: amputated", name)
				col = color.RGBA{100, 100, 100, 255}
			} else {
				pct := 0
				if part.MaxHP > 0 {
					pct = part.HP * 100 / part.MaxHP
					if pct < 0 {
						pct = 0
					}
				}
				label = fmt.Sprintf("%s: %d%%", name, pct)
				switch {
				case pct >= 50:
					col = color.RGBA{0, 255, 0, 255}
				case pct >= 25:
					col = color.RGBA{255, 255, 0, 255}
				default:
					col = color.RGBA{255, 0, 0, 255}
				}
			}
			mlge_text.Draw(screen, label, 14, config.Global().WorldWidth+4, y, col)
			y += 16
		}
	}

	g.initMsgArea()
	g.msgArea.Draw(screen)

	if config.Global().ShowMouseCoords {
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

			sw, sh := config.Global().SpriteSizeW, config.Global().SpriteSizeH
			if tile == nil {
				sX := 19 * sw
				worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+sw, sh)).(*ebiten.Image), op)
				continue
			}
			variant := cs.sim.Level.Level.ResolveVariant(tile)
			spriteX := variant.SpriteX * sw
			tx, ty, tz := tile.Coords()
			if !cs.sim.Level.GetSeen(tx, ty, tz) {
				spriteX = 19 * sw
			}
			worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(spriteX, variant.SpriteY, spriteX+sw, variant.SpriteY+sh)).(*ebiten.Image), op)
		}
	}

	ebitenutil.DrawRect(worldImage, float64(pc.GetX()*imageWidth/width), float64(pc.GetY()*imageHeight/height), 5, 5, color.RGBA{0, 0, 255, 255})

	return worldImage
}
