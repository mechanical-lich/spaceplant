package game

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// TermHUD is a static overlay showing HP and the recent message log.
// It is always visible and never consumes input.
type TermHUD struct {
	sim *SimWorld
}

func NewTermHUD(sim *SimWorld) *TermHUD        { return &TermHUD{sim: sim} }
func (h *TermHUD) Visible() bool               { return true }
func (h *TermHUD) HandleKey(*tcell.EventKey) bool { return false }

func (h *TermHUD) Draw(s tcell.Screen) {
	w, rows := s.Size()

	// HP and Z layer — top-right corner.
	if h.sim.Player != nil && h.sim.Player.HasComponent("Health") {
		hc := h.sim.Player.GetComponent("Health").(*component.HealthComponent)
		hp := fmt.Sprintf(" HP:%d/%d Z:%d ", hc.Health, hc.MaxHealth, h.sim.CurrentZ)
		rltermgui.DrawText(s, w-len(hp), 0, hp,
			tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack))
	}

	// Last 5 messages — bottom of screen.
	const maxMsgs = 5
	msgs := message.MessageLog
	start := len(msgs) - maxMsgs
	if start < 0 {
		start = 0
	}
	msgStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	for i, msg := range msgs[start:] {
		if len(msg) > w {
			msg = msg[:w]
		}
		rltermgui.DrawText(s, 0, rows-maxMsgs+i, msg, msgStyle)
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

	if v.sim.Player == nil || !v.sim.Player.HasComponent("Inventory") {
		rltermgui.DrawText(s, ix, iy, "No inventory.", dim)
		rltermgui.DrawText(s, ix, iy+ih-1, "[Esc] close", dim)
		return
	}
	inv := v.sim.Player.GetComponent("Inventory").(*component.InventoryComponent)

	colW := iw / 2

	// Left column: bag.
	rltermgui.DrawText(s, ix, iy, "── Bag ──", dim)
	if len(inv.Bag) == 0 {
		rltermgui.DrawText(s, ix, iy+1, "  (empty)", dim)
	} else {
		for i, item := range inv.Bag {
			if i >= ih-3 {
				break
			}
			name := itemName(item)
			line := fmt.Sprintf(" %d. %s", i+1, name)
			style := normal
			if i == v.cursor {
				style = selected
				// Pad so the highlight fills the column width.
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

	// Right column: equipped slots.
	rx := ix + colW + 1
	rltermgui.DrawText(s, rx, iy, "── Equipped ──", dim)
	slots := [6][2]string{
		{"Head   ", equippedName(inv.Head)},
		{"R.Hand ", equippedName(inv.RightHand)},
		{"L.Hand ", equippedName(inv.LeftHand)},
		{"Torso  ", equippedName(inv.Torso)},
		{"Legs   ", equippedName(inv.Legs)},
		{"Feet   ", equippedName(inv.Feet)},
	}
	maxRight := iw - colW - 1
	for i, sl := range slots {
		if i >= ih-3 {
			break
		}
		line := " " + sl[0] + sl[1]
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
		inv := v.inventoryOrNil()
		if inv != nil && v.cursor < len(inv.Bag)-1 {
			v.cursor++
		}
	}
	return true // consume all input while open
}

func (v *TermInventoryView) inventoryOrNil() *component.InventoryComponent {
	if v.sim.Player == nil || !v.sim.Player.HasComponent("Inventory") {
		return nil
	}
	return v.sim.Player.GetComponent("Inventory").(*component.InventoryComponent)
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
