package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
)

// WinConditionListener evaluates JSON-defined win/loss rules on game events.
type WinConditionListener struct {
	Sim SimAccess
}

func (l *WinConditionListener) HandleEvent(evt mlgeevent.EventData) error {
	switch evt.GetType() {
	case eventsystem.LifePodEscape:
		ctx := l.Sim.BuildEvalContext()
		if rule, ok := wincondition.Active().EvalInteraction("life_pod_escape", ctx); ok {
			wincondition.FireRule(rule, "")
		}

	case rlcomponents.DeathEventType:
		de, ok := evt.(rlcomponents.DeathEvent)
		if !ok {
			return nil
		}
		player := l.Sim.GetPlayer()
		if player == nil || de.Dying != player {
			return nil
		}
		ctx := l.Sim.BuildEvalContext()
		if rule, ok := wincondition.Active().EvalPlayerDeath(ctx); ok {
			wincondition.FireRule(rule, de.Message)
		}
	}
	return nil
}
