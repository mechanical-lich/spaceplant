package game

import (
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/world"

	"github.com/mechanical-lich/mlge/ecs"
)

// queuedMessageListener listens for queued message.MessageEvent and appends to MessageLog.
type queuedMessageListener struct {
	level  *world.Level
	player *ecs.Entity
}

func (q *queuedMessageListener) HandleEvent(evt mlgeevent.EventData) error {
	switch e := evt.(type) {
	case message.MessageEvent:
		// If the event has a location, only log it if it's visible to the player.
		if e.X != 0 || e.Y != 0 || e.Z != 0 {
			if q.level == nil || q.player == nil {
				break
			}
			pc := q.player.GetComponent("Position").(*component.PositionComponent)
			if pc.GetZ() != e.Z {
				return nil
			}
			if config.Los && !world.Los(pc.GetX(), pc.GetY(), e.X, e.Y, e.Z, q.level) {
				return nil
			}
		}
		message.AddMessage(e.Sender + ": " + e.Message)
	}
	return nil
}
