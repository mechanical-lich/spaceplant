package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

const (
	pauseMenuW = 220
	pauseMenuH = 304
)

// PauseMenu is shown when the player presses ESC during gameplay.
// It provides Save, Options, Controls, Return to Title, and Close (resume) options.
type PauseMenu struct {
	Visible  bool
	modal    *minui.Modal
	options  *OptionsModal
	controls *ControlsModal

	OnSave          func()
	OnReturnToTitle func()
}

func newPauseMenu() *PauseMenu {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - pauseMenuW) / 2
	my := (cfg.ScreenHeight - pauseMenuH) / 2

	pm := &PauseMenu{
		options:  newOptionsModal(),
		controls: newControlsModal(),
	}

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

	optionsBtn := minui.NewButton("pause_options", "Options")
	optionsBtn.SetPosition(btnX, 40+btnH+12)
	optionsBtn.SetSize(btnW, btnH)
	optionsBtn.OnClick = func() { pm.options.Open() }
	pm.modal.AddChild(optionsBtn)

	controlsBtn := minui.NewButton("pause_controls", "Controls")
	controlsBtn.SetPosition(btnX, 40+2*(btnH+12))
	controlsBtn.SetSize(btnW, btnH)
	controlsBtn.OnClick = func() { pm.controls.Open() }
	pm.modal.AddChild(controlsBtn)

	titleBtn := minui.NewButton("pause_title", "Return to Title")
	titleBtn.SetPosition(btnX, 40+3*(btnH+12))
	titleBtn.SetSize(btnW, btnH)
	titleBtn.OnClick = func() {
		pm.Visible = false
		if pm.OnReturnToTitle != nil {
			pm.OnReturnToTitle()
		}
	}
	pm.modal.AddChild(titleBtn)

	closeBtn := minui.NewButton("pause_close", "Close")
	closeBtn.SetPosition(btnX, 40+4*(btnH+12))
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
	if pm.options.Visible {
		pm.options.Update()
		return
	}
	if pm.controls.Visible {
		pm.controls.Update()
		return
	}
	pm.modal.Update()
}

func (pm *PauseMenu) Draw(screen *ebiten.Image) {
	if !pm.Visible {
		return
	}
	pm.modal.Draw(screen)
	pm.options.Draw(screen)
	pm.controls.Draw(screen)
}
