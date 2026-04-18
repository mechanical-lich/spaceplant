package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
)

// gameWonListener fires when a GameWonEventData is emitted.
type gameWonListener struct {
	modal *WinModal
}

func (l *gameWonListener) HandleEvent(evt event.EventData) error {
	e, ok := evt.(eventsystem.GameWonEventData)
	if !ok {
		return nil
	}
	l.modal.Show(e.Outcome, e.Message)
	return nil
}

const (
	winModalW = 420
	winModalH = 220
)

// WinModal is shown when the player achieves a win condition.
type WinModal struct {
	Visible         bool
	Outcome         string
	OnReturnToTitle func()
	modal           *minui.Modal
	msgLabel        *minui.Label
}

func newWinModal() *WinModal {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - winModalW) / 2
	my := (cfg.ScreenHeight - winModalH) / 2

	wm := &WinModal{}
	wm.modal = minui.NewModal("win_modal", "Mission Complete", winModalW, winModalH)
	wm.modal.SetPosition(mx, my)
	wm.modal.Closeable = false

	wm.msgLabel = minui.NewLabel("win_msg", "")
	wm.msgLabel.SetPosition(10, 10)
	wm.msgLabel.SetSize(winModalW-20, 100)
	wm.modal.AddChild(wm.msgLabel)

	btn := minui.NewButton("win_return", "Return to Title")
	btn.SetPosition((winModalW-160)/2, winModalH-90)
	btn.SetSize(160, 34)
	btn.OnClick = func() {
		if wm.OnReturnToTitle != nil {
			wm.OnReturnToTitle()
		}
	}
	wm.modal.AddChild(btn)

	return wm
}

func (wm *WinModal) Show(outcome, msg string) {
	wm.Outcome = outcome
	wm.msgLabel.Text = msg
	wm.Visible = true
}

func (wm *WinModal) Update() {
	if !wm.Visible {
		return
	}
	wm.modal.Update()
}

func (wm *WinModal) Draw(screen *ebiten.Image) {
	if !wm.Visible {
		return
	}
	wm.modal.Draw(screen)
}
