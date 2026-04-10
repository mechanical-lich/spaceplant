package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
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
	StatusText      string
	IsDoor          bool
	DoorOpen        bool
	DoorLocked      bool
	DoorKeyName     string
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
	minimapWidget *minui.ImageWidget
	msgArea       *minui.ScrollingTextArea
	hoverPanel    *minui.RichText
	msgSynced     int // number of MessageLog entries already pushed to msgArea
	hover         *hoveredTileInfo
}

func (g *GUIViewMain) initWidgets() {
	if g.minimapWidget != nil {
		return
	}
	cfg := config.Global()
	sideX := cfg.WorldWidth + 5
	sideW := cfg.ScreenWidth - cfg.WorldWidth - 8

	g.minimapWidget = minui.NewImageWidget("minimap", 150, 150)
	g.minimapWidget.SetPosition(sideX, 16)

	g.hoverPanel = minui.NewRichText("hover_panel", sideW)
	g.hoverPanel.LineHeight = 14

	g.msgArea = minui.NewScrollingTextArea("msglog", sideW, msgPanelH)
	g.msgArea.SetPosition(sideX-1, cfg.ScreenHeight-msgPanelH-4)
	g.msgArea.LineHeight = 15
}

func (g *GUIViewMain) Update(s any) {
	g.initWidgets()
	cs, ok := s.(*SPClientState)
	if ok {
		g.minimapWidget.Image = cs.GetMinimap(0, 0, 100, 100, 150, 150)
		g.updateHover(cs)
	}

	if g.hover != nil {
		cfg := config.Global()
		wrapChars := (cfg.ScreenWidth - cfg.WorldWidth - 12) / 7
		g.rebuildHoverSpans(wrapChars)
	} else {
		g.hoverPanel.Clear()
	}

	// Append any new messages (MessageLog grows monotonically).
	for g.msgSynced < len(message.MessageLog) {
		g.msgArea.AddText("> " + message.MessageLog[g.msgSynced])
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
		hasDesc := e.HasComponent(component.Description)
		hasItem := e.HasComponent(component.Item)
		if !hasDesc && !hasItem {
			continue
		}
		name := ""
		longDesc := ""
		if hasDesc {
			dc := e.GetComponent(component.Description).(*component.DescriptionComponent)
			name = dc.Name
			longDesc = dc.LongDescription
		}
		if hasItem {
			ic := e.GetComponent(component.Item).(*component.ItemComponent)
			if ic.Name != "" {
				name = ic.Name
			}
			if ic.Description != "" {
				longDesc = ic.Description
			}
		}
		ei := entityHoverInfo{
			Name:            name,
			LongDescription: longDesc,
		}
		if e.HasComponent(component.Door) {
			dc2 := e.GetComponent(component.Door).(*component.DoorComponent)
			ei.IsDoor = true
			ei.DoorOpen = dc2.Open
			ei.DoorLocked = dc2.Locked
			if dc2.Locked && dc2.KeyId != "" {
				ei.DoorKeyName = findKeyDisplayName(cs.sim, dc2.KeyId)
			}
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
		ei.StatusText = entityStatusText(e)
		info.Entities = append(info.Entities, ei)
	}
	g.hover = info
}

func (g *GUIViewMain) rebuildHoverSpans(wrapChars int) {
	g.hoverPanel.Clear()
	h := g.hover

	dimCol := color.RGBA{150, 150, 150, 255}
	whiteCol := color.RGBA{255, 255, 255, 255}
	headerCol := color.RGBA{200, 200, 100, 255}

	lightPct := (255 - h.LightLevel) * 100 / 255
	g.hoverPanel.AddSpan(minui.TextSpan{
		Text:  fmt.Sprintf("%s  (%d,%d,%d)", h.TileName, h.X, h.Y, h.Z),
		Color: headerCol,
		Size:  12,
	})
	g.hoverPanel.AddSpan(minui.TextSpan{
		Text:   fmt.Sprintf("light: %d%%", lightPct),
		Color:  dimCol,
		Size:   11,
		Indent: 4,
	})
	if h.TileDescription != "" {
		for _, line := range mlge_text.Wrap(h.TileDescription, wrapChars, 0) {
			g.hoverPanel.AddSpan(minui.TextSpan{Text: line, Color: dimCol, Size: 11, Indent: 4})
		}
	}
	g.hoverPanel.AddSpan(minui.TextSpan{Text: ""})

	for _, e := range h.Entities {
		g.hoverPanel.AddSpan(minui.TextSpan{Text: e.Name, Color: whiteCol, Size: 12})
		g.hoverPanel.AddSpan(minui.TextSpan{Text: "Status: " + e.StatusText, Color: dimCol, Size: 11, Indent: 4})
		if e.LongDescription != "" {
			for _, line := range mlge_text.Wrap(e.LongDescription, wrapChars, 0) {
				g.hoverPanel.AddSpan(minui.TextSpan{Text: line, Color: dimCol, Size: 11, Indent: 4})
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
			g.hoverPanel.AddSpan(minui.TextSpan{
				Text: fmt.Sprintf("%s  %s", doorState, lockState), Color: dimCol, Size: 11, Indent: 4,
			})
			if e.DoorKeyName != "" {
				g.hoverPanel.AddSpan(minui.TextSpan{
					Text: fmt.Sprintf("requires: %s", e.DoorKeyName), Color: dimCol, Size: 11, Indent: 4,
				})
			}
		}
		for _, p := range e.BodyParts {
			col, strike := bodyPartStyle(p)
			g.hoverPanel.AddSpan(minui.TextSpan{
				Text: p.Name, Color: col, Size: 11, Indent: 4, Strikethrough: strike,
			})
		}
		g.hoverPanel.AddSpan(minui.TextSpan{Text: ""})
	}
}

func (g *GUIViewMain) Draw(screen *ebiten.Image, s any) {
	g.initWidgets()
	g.minimapWidget.Draw(screen)

	cs, _ := s.(*SPClientState)
	if cs == nil {
		return
	}

	cfg := config.Global()
	mlge_text.Draw(screen,
		fmt.Sprintf("Turn: %d  Tick: %d", cs.sim.TurnCount, cs.sim.TickCount),
		12, cfg.WorldWidth+5, 170, color.RGBA{200, 200, 200, 255})

	y := 185
	if cs.sim.Player != nil && cs.sim.Player.HasComponent(rlcomponents.Energy) {
		ec := cs.sim.Player.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		var energyCol color.RGBA
		if ec.Energy < 0 {
			energyCol = color.RGBA{255, 60, 60, 255}
		} else {
			energyCol = color.RGBA{100, 200, 255, 255}
		}
		mlge_text.Draw(screen, fmt.Sprintf("Energy: %d", ec.Energy), 14, cfg.WorldWidth+4, y, energyCol)
		y += 20
	}

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
			mlge_text.Draw(screen, label, 14, cfg.WorldWidth+4, y, col)
			y += 16
		}
	}

	if cs.sim.Player != nil {
		statusText := entityStatusText(cs.sim.Player)
		mlge_text.Draw(screen, "Status: "+statusText, 14, cfg.WorldWidth+4, y, color.RGBA{200, 200, 200, 255})
		y += 18
	}

	if g.hover != nil {
		g.hoverPanel.SetPosition(cfg.WorldWidth+4, y+8)
		g.hoverPanel.Draw(screen)
	}

	g.msgArea.Draw(screen)

	if cfg.ShowMouseCoords {
		cX, cY := ebiten.CursorPosition()
		mlge_text.Draw(screen, strconv.Itoa(cX)+","+strconv.Itoa(cY), 16, cX, cY, color.RGBA{255, 0, 0, 255})
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
		return color.RGBA{0, 220, 0, 255}, false // green
	case pct >= 50:
		return color.RGBA{180, 220, 0, 255}, false // yellow-green
	case pct >= 25:
		return color.RGBA{255, 180, 0, 255}, false // orange
	default:
		return color.RGBA{220, 50, 50, 255}, false // red
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

// entityStatusText returns a comma-separated list of active status conditions
// with remaining durations, e.g. "poisoned (3), slowed (1)", or "normal".
func entityStatusText(e *ecs.Entity) string {
	type statusCheck struct {
		ct   ecs.ComponentType
		name string
	}
	checks := []statusCheck{
		{rlcomponents.Haste, "hasted"},
		{rlcomponents.Slowed, "slowed"},
		{rlcomponents.Alerted, "alerted"},
	}
	var active []string
	for _, sc := range checks {
		if !e.HasComponent(sc.ct) {
			continue
		}
		turns := 0
		switch v := e.GetComponent(sc.ct).(type) {
		case *rlcomponents.HasteComponent:
			turns = v.Duration
		case *rlcomponents.SlowedComponent:
			turns = v.Duration
		case *rlcomponents.AlertedComponent:
			turns = v.Duration
		}
		active = append(active, fmt.Sprintf("%s (%d)", sc.name, turns))
	}

	// Conditions stored in the ActiveConditionsComponent container.
	if e.HasComponent(rlcomponents.ActiveConditions) {
		acc := e.GetComponent(rlcomponents.ActiveConditions).(*rlcomponents.ActiveConditionsComponent)
		for _, d := range acc.Items {
			name := ""
			turns := 0
			type named interface{ GetConditionName() string }
			if n, ok := d.(named); ok {
				name = n.GetConditionName()
			}
			switch v := d.(type) {
			case *rlcomponents.DamageConditionComponent:
				turns = v.Duration
				if name == "" {
					name = "damage"
				}
			case *rlcomponents.StatConditionComponent:
				turns = v.Duration
				if name == "" {
					name = "condition"
				}
			default:
				if name == "" {
					name = "condition"
				}
			}
			active = append(active, fmt.Sprintf("%s (%d)", name, turns))
		}
	}

	if len(active) == 0 {
		return "normal"
	}
	return strings.Join(active, ", ")
}

// findKeyDisplayName searches level entities for a key matching keyID and
// returns its display name, falling back to keyID if none is found.
func findKeyDisplayName(sim *SimWorld, keyID string) string {
	for _, e := range sim.Level.Level.Entities {
		if e == nil || !e.HasComponent(rlcomponents.Key) {
			continue
		}
		kc := e.GetComponent(rlcomponents.Key).(*rlcomponents.KeyComponent)
		if kc.KeyID != keyID {
			continue
		}
		if e.HasComponent(component.Description) {
			return e.GetComponent(component.Description).(*component.DescriptionComponent).Name
		}
		return keyID
	}
	return keyID
}
