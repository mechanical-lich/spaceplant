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

	cm.input = minui.NewTextInput("cheat_input", "teleport <x> <y> <z>  |  heal <amount>")
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
		pc := cm.sim.Player.GetComponent("Position").(*component.PositionComponent)
		pc.SetPosition(x, y, z)
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
