package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

const (
	pauseMenuW = 220
	pauseMenuH = 210
)

// PauseMenu is shown when the player presses ESC during gameplay.
// It provides Save, Return to Title, and Close (resume) options.
type PauseMenu struct {
	Visible bool
	modal   *minui.Modal

	OnSave          func()
	OnReturnToTitle func()
}

func newPauseMenu() *PauseMenu {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - pauseMenuW) / 2
	my := (cfg.ScreenHeight - pauseMenuH) / 2

	pm := &PauseMenu{}

	pm.modal = minui.NewModal("pause_menu", "Paused", pauseMenuW, pauseMenuH)
	pm.modal.SetPosition(mx, my)
	pm.modal.Closeable = false

	btnW := 160
	btnH := 34
	btnX := (pauseMenuW - btnW) / 2

	saveBtn := minui.NewButton("pause_save", "Save")
	saveBtn.SetPosition(btnX, 40)
	saveBtn.SetSize(btnW, btnH)
	saveBtn.OnClick = func() {
		if pm.OnSave != nil {
			pm.OnSave()
		}
		pm.Visible = false
	}
	pm.modal.AddChild(saveBtn)

	titleBtn := minui.NewButton("pause_title", "Return to Title")
	titleBtn.SetPosition(btnX, 40+btnH+12)
	titleBtn.SetSize(btnW, btnH)
	titleBtn.OnClick = func() {
		pm.Visible = false
		if pm.OnReturnToTitle != nil {
			pm.OnReturnToTitle()
		}
	}
	pm.modal.AddChild(titleBtn)

	closeBtn := minui.NewButton("pause_close", "Close")
	closeBtn.SetPosition(btnX, 40+2*(btnH+12))
	closeBtn.SetSize(btnW, btnH)
	closeBtn.OnClick = func() { pm.Visible = false }
	pm.modal.AddChild(closeBtn)

	return pm
}

func (pm *PauseMenu) Open() { pm.Visible = true }

func (pm *PauseMenu) Update() {
	if !pm.Visible {
		return
	}
	pm.modal.Update()
}

func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	if !pm.Visible {
		return
	}
	pm.modal.Draw(screen)
}
