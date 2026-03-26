package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/ui"
)

const msgPanelH = 250

type bodyPartHoverInfo struct {
	Name      string
	HP, MaxHP int
	Broken    bool
	Amputated bool
}

type entityHoverInfo struct {
	Name            string
	LongDescription string
	BodyParts       []bodyPartHoverInfo
	IsDoor          bool
	DoorOpen        bool
	DoorLocked      bool
}

type hoveredTileInfo struct {
	TileName        string
	TileDescription string
	X, Y, Z         int
	LightLevel      int
	Entities        []entityHoverInfo
}

// Main gui
type GUIViewMain struct {
	ui.GUIViewBase
	minimap   *ebiten.Image
	msgArea   *minui.ScrollingTextArea
	msgSynced int // number of MessageLog entries already pushed to msgArea
	hover     *hoveredTileInfo
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
		g.updateHover(cs)
	}
	// Append any new messages (MessageLog grows monotonically).
	for g.msgSynced < len(message.MessageLog) {
		g.msgArea.AddText(message.MessageLog[g.msgSynced])
		g.msgSynced++
	}
	g.msgArea.Update()
}

func (g *GUIViewMain) updateHover(cs *SPClientState) {
	g.hover = nil
	cfg := config.Global()
	mx, my := ebiten.CursorPosition()
	if mx < 0 || mx >= cfg.WorldWidth || my < 0 || my >= cfg.WorldHeight {
		return
	}
	scale := math.Round(cfg.RenderScale)
	if scale < 1 {
		scale = 1
	}
	tilesW := int(math.Ceil(float64(cfg.WorldWidth) / (float64(cfg.SpriteSizeW) * scale)))
	tilesH := int(math.Ceil(float64(cfg.WorldHeight) / (float64(cfg.SpriteSizeH) * scale)))
	left := cs.CameraX - tilesW/2
	up := cs.CameraY - tilesH/2
	tileX := left + int(float64(mx)/scale)/cfg.SpriteSizeW
	tileY := up + int(float64(my)/scale)/cfg.SpriteSizeH

	tile := cs.sim.Level.Level.GetTilePtr(tileX, tileY, cs.sim.CurrentZ)
	if tile == nil {
		return
	}
	def := rlworld.TileDefinitions[tile.Type]
	info := &hoveredTileInfo{
		TileName:        def.Name,
		TileDescription: def.Description,
		X:               tileX,
		Y:               tileY,
		Z:               cs.sim.CurrentZ,
		LightLevel:      tile.LightLevel,
	}
	var buf []*ecs.Entity
	cs.sim.Level.Level.GetEntitiesAt(tileX, tileY, cs.sim.CurrentZ, &buf)
	for _, e := range buf {
		if !e.HasComponent(component.Description) {
			continue
		}
		dc := e.GetComponent(component.Description).(*component.DescriptionComponent)
		ei := entityHoverInfo{
			Name:            dc.Name,
			LongDescription: dc.LongDescription,
		}
		if e.HasComponent(component.Door) {
			dc2 := e.GetComponent(component.Door).(*component.DoorComponent)
			ei.IsDoor = true
			ei.DoorOpen = dc2.Open
			ei.DoorLocked = dc2.Locked
		}
		if e.HasComponent(component.Body) {
			bc := e.GetComponent(component.Body).(*component.BodyComponent)
			partKeys := make([]string, 0, len(bc.Parts))
			for k := range bc.Parts {
				partKeys = append(partKeys, k)
			}
			slices.Sort(partKeys)
			for _, pk := range partKeys {
				p := bc.Parts[pk]
				ei.BodyParts = append(ei.BodyParts, bodyPartHoverInfo{
					Name:      p.Name,
					HP:        p.HP,
					MaxHP:     p.MaxHP,
					Broken:    p.Broken,
					Amputated: p.Amputated,
				})
			}
		}
		info.Entities = append(info.Entities, ei)
	}
	g.hover = info
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

	y := 185
	if cs.sim.Player != nil && cs.sim.Player.HasComponent(component.Body) {
		bc := cs.sim.Player.GetComponent(component.Body).(*component.BodyComponent)
		keys := make([]string, 0, len(bc.Parts))
		for k := range bc.Parts {
			keys = append(keys, k)
		}
		slices.Sort(keys)
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

	if g.hover != nil {
		g.drawHoverPanel(screen, y+8)
	}

	g.initMsgArea()
	g.msgArea.Draw(screen)

	if config.Global().ShowMouseCoords {
		cX, cY := ebiten.CursorPosition()
		mlge_text.Draw(screen, strconv.Itoa(cX)+","+strconv.Itoa(cY), 16, cX, cY, color.RGBA{255, 0, 0, 255})
	}
}

func (g *GUIViewMain) drawHoverPanel(screen *ebiten.Image, startY int) {
	cfg := config.Global()
	x := cfg.WorldWidth + 4
	wrapChars := (cfg.ScreenWidth - cfg.WorldWidth - 12) / 7

	dimCol := color.RGBA{150, 150, 150, 255}
	whiteCol := color.RGBA{255, 255, 255, 255}
	headerCol := color.RGBA{200, 200, 100, 255}

	y := startY
	lineH := 15

	h := g.hover
	lightPct := (255 - h.LightLevel) * 100 / 255
	mlge_text.Draw(screen, fmt.Sprintf("%s  (%d,%d,%d)", h.TileName, h.X, h.Y, h.Z), 12, x, y, headerCol)
	y += lineH
	mlge_text.Draw(screen, fmt.Sprintf("light: %d%%", lightPct), 11, x+4, y, dimCol)
	y += lineH - 1
	if h.TileDescription != "" {
		for _, line := range mlge_text.Wrap(h.TileDescription, wrapChars, 0) {
			mlge_text.Draw(screen, line, 11, x+4, y, dimCol)
			y += lineH - 1
		}
	}
	y += 2

	for _, e := range h.Entities {
		mlge_text.Draw(screen, e.Name, 12, x, y, whiteCol)
		y += lineH
		if e.LongDescription != "" {
			for _, line := range mlge_text.Wrap(e.LongDescription, wrapChars, 0) {
				mlge_text.Draw(screen, line, 11, x+4, y, dimCol)
				y += lineH - 1
			}
		}
		if e.IsDoor {
			doorState := "closed"
			if e.DoorOpen {
				doorState = "open"
			}
			lockState := "unlocked"
			if e.DoorLocked {
				lockState = "locked"
			}
			mlge_text.Draw(screen, fmt.Sprintf("%s  %s", doorState, lockState), 11, x+4, y, dimCol)
			y += lineH - 1
		}
		for _, p := range e.BodyParts {
			col, strike := bodyPartStyle(p)
			const partSize = 11
			mlge_text.Draw(screen, p.Name, partSize, x+4, y, col)
			if strike {
				w, _ := mlge_text.Measure(p.Name, partSize)
				ebitenutil.DrawRect(screen, float64(x+4), float64(y)+float64(partSize)/2, w, 1, col)
			}
			y += lineH - 2
		}
		y += 3
	}
}

// bodyPartStyle returns the draw colour and whether to strike through a body part label.
func bodyPartStyle(p bodyPartHoverInfo) (color.RGBA, bool) {
	if p.Amputated {
		return color.RGBA{120, 80, 160, 255}, true // purple + strikethrough
	}
	if p.Broken {
		return color.RGBA{160, 80, 200, 255}, false // purple, no strikethrough
	}
	pct := 100
	if p.MaxHP > 0 {
		pct = p.HP * 100 / p.MaxHP
		if pct < 0 {
			pct = 0
		}
	}
	switch {
	case pct >= 75:
		return color.RGBA{0, 220, 0, 255}, false   // green
	case pct >= 50:
		return color.RGBA{180, 220, 0, 255}, false  // yellow-green
	case pct >= 25:
		return color.RGBA{255, 180, 0, 255}, false  // orange
	default:
		return color.RGBA{220, 50, 50, 255}, false  // red
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
