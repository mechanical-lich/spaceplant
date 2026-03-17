package eventsystem

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

var EventManager *event.QueuedEventManager

const (
	Stairs   event.EventType = "Stairs"
	DropItem event.EventType = "DropItem"
)

func init() {
	EventManager = &event.QueuedEventManager{}
}

type StairsEventData struct {
	Up bool
}

func (t StairsEventData) GetType() event.EventType {
	return Stairs
}

type DropItemEventData struct {
	X, Y, Z int
	Item *ecs.Entity
}

func (t DropItemEventData) GetType() event.EventType {
	return DropItem
}
