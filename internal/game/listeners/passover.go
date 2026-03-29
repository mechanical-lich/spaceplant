package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// PassoverListener handles PassoverEvents and logs a message only when the
// player was the one who moved onto the tile.
type PassoverListener struct {
	Sim SimAccess
}

func (l *PassoverListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(rlcomponents.PassoverEvent)
	if !ok {
		return nil
	}
	player := l.Sim.GetPlayer()
	if player == nil || e.Mover != player {
		return nil
	}
	message.AddMessage(rlentity.GetName(e.SteppedOn) + ": " + e.Message)
	return nil
}
