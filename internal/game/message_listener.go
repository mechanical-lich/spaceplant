package game

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// queuedMessageListener listens for queued message.MessageEvent and appends to MessageLog.
type queuedMessageListener struct {
	sim *SimWorld
}

func (q *queuedMessageListener) HandleEvent(evt mlgeevent.EventData) error {
	switch e := evt.(type) {
	case message.MessageEvent:
		// If the event has a location, only log it if it's visible to the player.
		if e.X != 0 || e.Y != 0 || e.Z != 0 {
			if q.sim.Level == nil || q.sim.Player == nil {
				break
			}
			pc := q.sim.Player.GetComponent("Position").(*component.PositionComponent)
			if pc.GetZ() != e.Z {
				return nil
			}
			if config.Global().Los && !rlfov.Los(q.sim.Level.Level, pc.GetX(), pc.GetY(), e.X, e.Y, e.Z) {
				return nil
			}
		}
		message.AddMessage(e.Sender + ": " + e.Message)
	}
	return nil
}
