package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// playerDeathListener fires when the player's own death event arrives.
type playerDeathListener struct {
	player *ecs.Entity
	modal  *DeathModal
}

func (l *playerDeathListener) HandleEvent(evt mlgeevent.EventData) error {
	de, ok := evt.(rlcomponents.DeathEvent)
	if !ok || de.Dying != l.player {
		return nil
	}
	l.modal.Show(de.Message)
	return nil
}

const (
	deathModalW = 380
	deathModalH = 200
)

// DeathModal is shown when the player dies. It displays the cause of death
// and a button to return to the title screen.
type DeathModal struct {
	Visible bool
	modal   *minui.Modal
	msgLabel *minui.Label

	OnReturnToTitle func()
}

func newDeathModal() *DeathModal {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - deathModalW) / 2
	my := (cfg.ScreenHeight - deathModalH) / 2

	dm := &DeathModal{}

	dm.modal = minui.NewModal("death_modal", "You have died.", deathModalW, deathModalH)
	dm.modal.SetPosition(mx, my)
	dm.modal.Closeable = false

	dm.msgLabel = minui.NewLabel("death_msg", "")
	dm.msgLabel.SetPosition(10, 10)
	dm.msgLabel.SetSize(deathModalW-20, 80)
	dm.modal.AddChild(dm.msgLabel)

	btn := minui.NewButton("death_return", "Return to Title")
	btn.SetPosition((deathModalW-160)/2, deathModalH-90)
	btn.SetSize(160, 34)
	btn.OnClick = func() {
		if dm.OnReturnToTitle != nil {
			dm.OnReturnToTitle()
		}
	}
	dm.modal.AddChild(btn)

	return dm
}

// Show displays the modal with the given death message.
func (dm *DeathModal) Show(msg string) {
	dm.msgLabel.Text = msg
	dm.Visible = true
}

func (dm *DeathModal) Update() {
	if !dm.Visible {
		return
	}
	dm.modal.Update()
}

func (dm *DeathModal) Draw(screen *ebiten.Image) {
	if !dm.Visible {
		return
	}
	dm.modal.Draw(screen)
}
