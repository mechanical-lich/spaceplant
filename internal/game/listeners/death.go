package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
)

// DeathListener handles DeathEvents and logs a message only when the player
// witnessed the death. LOS is already checked before the event is queued.
// It also detects mother plant death and fires GameWon.
type DeathListener struct {
	Sim SimAccess
}

func (l *DeathListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(rlcomponents.DeathEvent)
	if !ok {
		return nil
	}

	if e.Dying != nil && e.Dying.Blueprint == "mother_plant" {
		eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{
			Outcome: "extermination",
			Message: "The mother plant is dead. The infestation collapses.",
		})
		return nil
	}

	player := l.Sim.GetPlayer()
	if player == nil || e.Watcher != player {
		return nil
	}
	message.AddMessage(e.Message)
	return nil
}
