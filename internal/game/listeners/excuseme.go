package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// ExcuseMeListener handles ExcuseMeEvents and logs a message only when the
// player was the one doing the bumping.
type ExcuseMeListener struct {
	Sim SimAccess
}

func (l *ExcuseMeListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(rlcomponents.ExcuseMeEvent)
	if !ok {
		return nil
	}
	player := l.Sim.GetPlayer()
	if player == nil || e.Mover != player {
		return nil
	}
	message.AddMessage(rlentity.GetName(e.Bumped) + ": " + e.Message)
	return nil
}
