package game

import (
	"fmt"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

const termMsgLines = 8 // visible message rows at the bottom

// TermHUD is a static overlay showing body-part health and the recent message log.
// PgUp/PgDn scroll the message log; all other keys pass through to the game.
type TermHUD struct {
	sim         *SimWorld
	msgOffset   int // 0 = newest at bottom; positive = scrolled back in history
}

func NewTermHUD(sim *SimWorld) *TermHUD { return &TermHUD{sim: sim} }
func (h *TermHUD) Visible() bool        { return true }

func (h *TermHUD) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyPgUp:
		h.msgOffset += termMsgLines
		return true
	case tcell.KeyPgDn:
		h.msgOffset -= termMsgLines
		if h.msgOffset < 0 {
			h.msgOffset = 0
		}
		return true
	}
	return false
}

func (h *TermHUD) Draw(s tcell.Screen) {
	w, rows := s.Size()

	// Body part health — stacked in top-right corner.
	if h.sim.Player != nil && h.sim.Player.HasComponent(component.Body) {
		bc := h.sim.Player.GetComponent(component.Body).(*component.BodyComponent)
		keys := make([]string, 0, len(bc.Parts))
		for k := range bc.Parts {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for i, name := range keys {
			part := bc.Parts[name]
			var text string
			var fg tcell.Color
			if part.Amputated {
				text = fmt.Sprintf(" %s:amp ", name)
				fg = tcell.ColorGray
			} else {
				pct := 0
				if part.MaxHP > 0 {
					pct = part.HP * 100 / part.MaxHP
					if pct < 0 {
						pct = 0
					}
				}
				text = fmt.Sprintf(" %s:%d%% ", name, pct)
				switch {
				case pct >= 50:
					fg = tcell.ColorGreen
				case pct >= 25:
					fg = tcell.ColorYellow
				default:
					fg = tcell.ColorRed
				}
			}
			rltermgui.DrawText(s, w-len(text), i, text,
				tcell.StyleDefault.Foreground(fg).Background(tcell.ColorBlack))
		}
	}

	// Message log — bottom termMsgLines rows, scrollable with PgUp/PgDn.
	msgs := message.MessageLog
	total := len(msgs)

	// Clamp offset so we never scroll past the oldest message.
	maxOffset := total - termMsgLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if h.msgOffset > maxOffset {
		h.msgOffset = maxOffset
	}

	// The window end is "offset from the newest": offset 0 means the last
	// termMsgLines messages; offset N means N messages further back.
	endIdx := total - h.msgOffset
	startIdx := endIdx - termMsgLines
	if startIdx < 0 {
		startIdx = 0
	}

	msgStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	dimStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)

	for i, msg := range msgs[startIdx:endIdx] {
		if len([]rune(msg)) > w {
			msg = string([]rune(msg)[:w])
		}
		rltermgui.DrawText(s, 0, rows-termMsgLines+i, msg, msgStyle)
	}

	// Scroll hint when there are messages above or below the current window.
	hint := ""
	if h.msgOffset > 0 {
		hint += "↓PgDn "
	}
	if h.msgOffset < maxOffset {
		hint += "↑PgUp"
	}
	if hint != "" {
		rltermgui.DrawText(s, 0, rows-termMsgLines-1, hint, dimStyle)
	}
}

// TermInventoryView is a modal inventory panel for the terminal client.
// Hidden by default; call Show() or Toggle() to open it.
// All key events are consumed while open; Escape closes it.
type TermInventoryView struct {
	*rltermgui.Pane
	sim    *SimWorld
	cursor int
}

func NewTermInventoryView(sim *SimWorld) *TermInventoryView {
	p := rltermgui.NewPane(0, 0, 60, 24)
	p.Title = " Inventory "
	p.BorderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	p.ContentStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	return &TermInventoryView{Pane: p, sim: sim}
}

func (v *TermInventoryView) Draw(s tcell.Screen) {
	// Re-center every draw so it stays correct after terminal resize.
	sw, sh := s.Size()
	v.W = sw * 2 / 3
	if v.W < 50 {
		v.W = 50
	}
	v.H = sh * 2 / 3
	if v.H < 14 {
		v.H = 14
	}
	v.X = (sw - v.W) / 2
	v.Y = (sh - v.H) / 2

	v.DrawPane(s)
	ix, iy, iw, ih := v.Inner()

	normal := v.ContentStyle
	selected := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	dim := v.ContentStyle.Foreground(tcell.ColorGray)

	bag := playerBag(v.sim.Player)
	equipped := playerEquipped(v.sim.Player)

	if bag == nil && equipped == nil {
		rltermgui.DrawText(s, ix, iy, "No inventory.", dim)
		rltermgui.DrawText(s, ix, iy+ih-1, "[Esc] close", dim)
		return
	}
	if bag == nil {
		bag = nil // keep nil safe for range below
	}

	colW := iw / 2

	// Left column: bag.
	rltermgui.DrawText(s, ix, iy, "── Bag ──", dim)
	if len(bag) == 0 {
		rltermgui.DrawText(s, ix, iy+1, "  (empty)", dim)
	} else {
		for i, item := range bag {
			if i >= ih-3 {
				break
			}
			name := itemName(item)
			line := fmt.Sprintf(" %d. %s", i+1, name)
			style := normal
			if i == v.cursor {
				style = selected
				for len([]rune(line)) < colW {
					line += " "
				}
			}
			if len([]rune(line)) > colW {
				line = string([]rune(line)[:colW])
			}
			rltermgui.DrawText(s, ix, iy+1+i, line, style)
		}
	}

	// Right column: equipped slots (sorted by part name).
	rx := ix + colW + 1
	rltermgui.DrawText(s, rx, iy, "── Equipped ──", dim)
	eqKeys := make([]string, 0, len(equipped))
	for k := range equipped {
		eqKeys = append(eqKeys, k)
	}
	slices.Sort(eqKeys)
	maxRight := iw - colW - 1
	for i, slot := range eqKeys {
		if i >= ih-3 {
			break
		}
		line := fmt.Sprintf(" %-10s %s", slot, equippedName(equipped[slot]))
		if len([]rune(line)) > maxRight {
			line = string([]rune(line)[:maxRight])
		}
		rltermgui.DrawText(s, rx, iy+1+i, line, normal)
	}

	rltermgui.DrawText(s, ix, iy+ih-1, "[↑↓] select  [Esc] close", dim)
}

func (v *TermInventoryView) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		v.Hide()
	case tcell.KeyUp:
		if v.cursor > 0 {
			v.cursor--
		}
	case tcell.KeyDown:
		bag := playerBag(v.sim.Player)
		if v.cursor < len(bag)-1 {
			v.cursor++
		}
	}
	return true // consume all input while open
}

// TermLookView is a tile-inspection mode activated by pressing L.
// Arrow keys / WASD move the cursor; Escape exits. All input is consumed
// while active so the game does not act on movement keys.
type TermLookView struct {
	sim     *SimWorld
	active  bool
	cursorX int
	cursorY int
}

func NewTermLookView(sim *SimWorld) *TermLookView { return &TermLookView{sim: sim} }
func (v *TermLookView) Visible() bool             { return true } // always present to intercept L

func (v *TermLookView) HandleKey(ev *tcell.EventKey) bool {
	if !v.active {
		if ev.Rune() == 'l' || ev.Rune() == 'L' {
			v.active = true
			if v.sim.Player != nil {
				pc := v.sim.Player.GetComponent("Position").(*component.PositionComponent)
				v.cursorX = pc.GetX()
				v.cursorY = pc.GetY()
			}
			return true
		}
		return false
	}
	// Active — handle cursor movement and exit.
	if ev.Key() == tcell.KeyEscape || ev.Rune() == 'l' || ev.Rune() == 'L' {
		v.active = false
		return true
	}
	switch ev.Key() {
	case tcell.KeyUp:
		v.cursorY--
		return true
	case tcell.KeyDown:
		v.cursorY++
		return true
	case tcell.KeyLeft:
		v.cursorX--
		return true
	case tcell.KeyRight:
		v.cursorX++
		return true
	}
	switch ev.Rune() {
	case 'w', 'W':
		v.cursorY--
		return true
	case 's', 'S':
		v.cursorY++
		return true
	case 'a', 'A':
		v.cursorX--
		return true
	case 'd', 'D':
		v.cursorX++
		return true
	}
	return true // swallow all other input while in look mode
}

func (v *TermLookView) Draw(s tcell.Screen) {
	if !v.active {
		return
	}
	sw, sh := s.Size()

	// Recompute camera the same way OnTick does.
	cameraX, cameraY := 0, 0
	if v.sim.Player != nil {
		pc := v.sim.Player.GetComponent("Position").(*component.PositionComponent)
		cameraX = pc.GetX() - sw/2
		cameraY = pc.GetY() - sh/2
	}

	// Highlight cursor cell by inverting colours.
	sx := v.cursorX - cameraX
	sy := v.cursorY - cameraY
	if sx >= 0 && sx < sw && sy >= 0 && sy < sh {
		mainC, combC, _, _ := s.GetContent(sx, sy)
		s.SetContent(sx, sy, mainC, combC,
			tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow))
	}

	// Info panel — right side, above message rows.
	const panelW = 40
	panelH := sh - termMsgLines - 1
	if panelH < 4 {
		return
	}
	panelX := sw - panelW
	panelY := 0

	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	bgStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	headerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	dimStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)

	rltermgui.FillRect(s, panelX, panelY, panelW, panelH, bgStyle)
	rltermgui.DrawBox(s, panelX, panelY, panelW, panelH, " Look [L/Esc] ", borderStyle)

	ix := panelX + 1
	iw := panelW - 2
	maxY := panelY + panelH - 1
	y := panelY + 1

	drawLine := func(text string, style tcell.Style) bool {
		if y >= maxY {
			return false
		}
		if len([]rune(text)) > iw {
			text = string([]rune(text)[:iw])
		}
		rltermgui.DrawText(s, ix, y, text, style)
		y++
		return true
	}

	tile := v.sim.Level.Level.GetTilePtr(v.cursorX, v.cursorY, v.sim.CurrentZ)
	if tile == nil {
		drawLine(fmt.Sprintf("(%d,%d) — out of bounds", v.cursorX, v.cursorY), dimStyle)
		return
	}

	def := rlworld.TileDefinitions[tile.Type]
	lightPct := (255 - tile.LightLevel) * 100 / 255
	if !drawLine(fmt.Sprintf("%s (%d,%d) light:%d%%", def.Name, v.cursorX, v.cursorY, lightPct), headerStyle) {
		return
	}
	if def.Description != "" {
		for _, line := range mlge_text.Wrap(def.Description, iw, 0) {
			if !drawLine("  "+line, dimStyle) {
				return
			}
		}
	}
	y++ // gap before entities

	var buf []*ecs.Entity
	v.sim.Level.Level.GetEntitiesAt(v.cursorX, v.cursorY, v.sim.CurrentZ, &buf)
	for _, e := range buf {
		if !e.HasComponent(component.Description) {
			continue
		}
		dc := e.GetComponent(component.Description).(*component.DescriptionComponent)
		if !drawLine(dc.Name, bgStyle) {
			return
		}
		if dc.LongDescription != "" {
			for _, line := range mlge_text.Wrap(dc.LongDescription, iw-2, 0) {
				if !drawLine("  "+line, dimStyle) {
					return
				}
			}
		}
		if e.HasComponent(component.Door) {
			door := e.GetComponent(component.Door).(*component.DoorComponent)
			doorState := "closed"
			if door.Open {
				doorState = "open"
			}
			lockState := "unlocked"
			if door.Locked {
				lockState = "locked"
			}
			if !drawLine(fmt.Sprintf("  %s  %s", doorState, lockState), dimStyle) {
				return
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
				col, strike := termBodyPartStyle(p)
				st := tcell.StyleDefault.Foreground(col).Background(tcell.ColorBlack)
				if strike {
					st = st.Attributes(tcell.AttrStrikeThrough)
				}
				if !drawLine("  "+p.Name, st) {
					return
				}
			}
		}
		y++ // gap between entities
	}
}

func termBodyPartStyle(p component.BodyPart) (tcell.Color, bool) {
	if p.Amputated {
		return tcell.ColorPurple, true
	}
	if p.Broken {
		return tcell.ColorPurple, false
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
		return tcell.ColorGreen, false
	case pct >= 50:
		return tcell.ColorYellow, false
	case pct >= 25:
		return tcell.ColorOrange, false
	default:
		return tcell.ColorRed, false
	}
}

func itemName(e *ecs.Entity) string {
	if e != nil && e.HasComponent("Description") {
		return e.GetComponent("Description").(*component.DescriptionComponent).Name
	}
	return "?"
}

func equippedName(e *ecs.Entity) string {
	if e == nil {
		return "-"
	}
	return itemName(e)
}
