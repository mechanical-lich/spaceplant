package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// MessageListener handles queued message.MessageEvent and appends to MessageLog.
type MessageListener struct {
	Sim SimAccess
}

func (q *MessageListener) HandleEvent(evt mlgeevent.EventData) error {
	switch e := evt.(type) {
	case message.MessageEvent:
		// If the event has a location, only log it if it's visible to the player.
		if e.X != 0 || e.Y != 0 || e.Z != 0 {
			player := q.Sim.GetPlayer()
			level := q.Sim.GetRLLevel()
			if level == nil || player == nil {
				break
			}
			pc := player.GetComponent("Position").(*component.PositionComponent)
			if pc.GetZ() != e.Z {
				return nil
			}
			if config.Global().Los && !rlfov.Los(level, pc.GetX(), pc.GetY(), e.X, e.Y, e.Z) {
				return nil
			}
		}
		message.AddMessage(e.Sender + ": " + e.Message)
	}
	return nil
}
