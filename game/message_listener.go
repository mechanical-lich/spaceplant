package game

import (
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// queuedMessageListener listens for queued message.MessageEvent and appends to MessageLog.
type queuedMessageListener struct{}

func (q *queuedMessageListener) HandleEvent(evt mlgeevent.EventData) error {
	switch e := evt.(type) {
	case message.MessageEvent:
		// Append sender + message to the global MessageLog
		message.AddMessage(e.Sender + ": " + e.Message)
	}
	return nil
}
