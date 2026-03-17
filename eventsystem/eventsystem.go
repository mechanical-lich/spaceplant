package eventsystem

import (
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/game-engine/event"
)

var EventManager *event.QueuedEventManager

const (
	Stairs event.EventType = iota
	DropItem
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
	X, Y int
	Item *ecs.Entity
}

func (t DropItemEventData) GetType() event.EventType {
	return DropItem
}
