package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

const (
	cheatModalW = 400
	cheatModalH = 160
)

// CheatModal is a developer console triggered by Shift+ESC.
type CheatModal struct {
	Visible bool
	sim     *SimWorld
	modal   *minui.Modal
	input   *minui.TextInput
}

func newCheatModal(sim *SimWorld) *CheatModal {
	cm := &CheatModal{sim: sim}
	cm.rebuildModal()
	return cm
}

// Open resets and shows the modal with a fresh TextInput (avoids stale cursorPos).
func (cm *CheatModal) Open() {
	cm.rebuildModal()
	cm.Visible = true
}

func (cm *CheatModal) rebuildModal() {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - cheatModalW) / 2
	my := (cfg.ScreenHeight - cheatModalH) / 2

	cm.modal = minui.NewModal("cheat_modal", "Developer Console", cheatModalW, cheatModalH)
	cm.modal.SetPosition(mx, my)
	cm.modal.Closeable = false

	cm.input = minui.NewTextInput("cheat_input", "tp <x> <y> <z>  |  heal <n>  |  find <blueprint>  |  fr <room_tag>")
	cm.input.SetPosition(10, 20)
	cm.input.SetSize(cheatModalW-20, 30)
	cm.modal.AddChild(cm.input)

	okBtn := minui.NewButton("cheat_ok", "OK")
	okBtn.SetPosition((cheatModalW-180)/2-5, cheatModalH-75)
	okBtn.SetSize(80, 30)
	okBtn.OnClick = func() {
		cm.runCommand(cm.input.Text)
		cm.Visible = false
	}
	cm.modal.AddChild(okBtn)

	cancelBtn := minui.NewButton("cheat_cancel", "Cancel")
	cancelBtn.SetPosition((cheatModalW-180)/2+85, cheatModalH-75)
	cancelBtn.SetSize(80, 30)
	cancelBtn.OnClick = func() {
		cm.Visible = false
	}
	cm.modal.AddChild(cancelBtn)
}

func (cm *CheatModal) runCommand(raw string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return
	}
	switch strings.ToLower(parts[0]) {
	case "teleport", "tp":
		if len(parts) < 4 {
			message.AddMessage("[cheat] usage: teleport <x> <y> <z>")
			return
		}
		x, ex := strconv.Atoi(parts[1])
		y, ey := strconv.Atoi(parts[2])
		z, ez := strconv.Atoi(parts[3])
		if ex != nil || ey != nil || ez != nil {
			message.AddMessage("[cheat] teleport: invalid coordinates")
			return
		}
		if cm.sim.Player == nil {
			return
		}
		cm.sim.Mu.Lock()
		cm.sim.Level.Level.PlaceEntity(x, y, z, cm.sim.Player)
		cm.sim.CurrentZ = z
		cm.sim.UpdateEntities()
		cm.sim.Mu.Unlock()
		message.AddMessage(fmt.Sprintf("[cheat] teleported to %d,%d,%d", x, y, z))

	case "heal":
		amount := 50
		if len(parts) >= 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				amount = v
			}
		}
		if cm.sim.Player == nil {
			return
		}
		cm.sim.Mu.Lock()
		if cm.sim.Player.HasComponent("Body") {
			body := cm.sim.Player.GetComponent("Body")
			type healer interface{ Heal(int) }
			if h, ok := body.(healer); ok {
				h.Heal(amount)
			}
		}
		if cm.sim.Player.HasComponent("StatsComponent") {
			type hpSetter interface {
				GetHP() int
				SetHP(int)
				GetMaxHP() int
			}
			if sc, ok := cm.sim.Player.GetComponent("StatsComponent").(hpSetter); ok {
				hp := sc.GetHP() + amount
				if hp > sc.GetMaxHP() {
					hp = sc.GetMaxHP()
				}
				sc.SetHP(hp)
			}
		}
		cm.sim.Mu.Unlock()
		message.AddMessage(fmt.Sprintf("[cheat] healed %d HP", amount))

	case "find":
		if len(parts) < 2 {
			message.AddMessage("[cheat] usage: find <blueprint>")
			return
		}
		target := strings.ToLower(parts[1])
		if cm.sim.Player == nil {
			return
		}
		pc := cm.sim.Player.GetComponent("Position").(*component.PositionComponent)
		px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()

		bestDist := -1
		bestX, bestY, bestZ := 0, 0, 0
		all := append(cm.sim.Level.Level.GetEntities(), cm.sim.Level.Level.GetStaticEntities()...)
		for _, e := range all {
			if e == nil || !strings.EqualFold(e.Blueprint, target) {
				continue
			}
			if !e.HasComponent("Position") {
				continue
			}
			epc := e.GetComponent("Position").(*component.PositionComponent)
			dx := epc.GetX() - px
			dy := epc.GetY() - py
			dz := (epc.GetZ() - pz) * 50
			dist := dx*dx + dy*dy + dz*dz
			if bestDist < 0 || dist < bestDist {
				bestDist = dist
				bestX, bestY, bestZ = epc.GetX(), epc.GetY(), epc.GetZ()
			}
		}
		if bestDist < 0 {
			message.AddMessage(fmt.Sprintf("[cheat] no %q found", target))
		} else {
			message.AddMessage(fmt.Sprintf("[cheat] nearest %q at %d,%d,%d", target, bestX, bestY, bestZ))
		}

	case "find_room", "fr":
		if len(parts) < 2 {
			message.AddMessage("[cheat] usage: find_room <tag>")
			return
		}
		target := strings.ToLower(parts[1])
		if cm.sim.Player == nil {
			return
		}
		pc := cm.sim.Player.GetComponent("Position").(*component.PositionComponent)
		px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()

		bestDist := -1
		bestX, bestY, bestZ := 0, 0, 0
		for _, fr := range cm.sim.FloorResults {
			for _, room := range fr.Rooms {
				if !strings.EqualFold(room.Tag, target) {
					continue
				}
				dx := room.X - px
				dy := room.Y - py
				dz := (fr.Z - pz) * 50
				dist := dx*dx + dy*dy + dz*dz
				if bestDist < 0 || dist < bestDist {
					bestDist = dist
					bestX, bestY, bestZ = room.X, room.Y, fr.Z
				}
			}
		}
		if bestDist < 0 {
			message.AddMessage(fmt.Sprintf("[cheat] no room %q found", target))
		} else {
			message.AddMessage(fmt.Sprintf("[cheat] nearest %q at %d,%d,%d", target, bestX, bestY, bestZ))
		}

	default:
		message.AddMessage("[cheat] unknown command: " + parts[0])
	}
}

func (cm *CheatModal) Update() {
	if !cm.Visible {
		return
	}
	cm.modal.Update()
}

func (cm *CheatModal) Draw(screen *ebiten.Image) {
	if !cm.Visible {
		return
	}
	cm.modal.Draw(screen)
}
