package eventsystem

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

var EventManager *event.QueuedEventManager

const (
	Stairs          event.EventType = "Stairs"
	DropItem        event.EventType = "DropItem"
	GameWon         event.EventType = "GameWon"
	PlaceMotherPlant event.EventType = "PlaceMotherPlant"
	ArmSelfDestruct  event.EventType = "ArmSelfDestruct"
	LifePodEscape    event.EventType = "LifePodEscape"
)

func init() {
	EventManager = &event.QueuedEventManager{}
}

// GameWonEventData is emitted by any system that determines the player has won.
// Outcome is a short identifier (e.g. "escape_selfish", "saboteur", "extermination").
type GameWonEventData struct {
	Outcome string
	Message string
}

func (e GameWonEventData) GetType() event.EventType { return GameWon }

// PlaceMotherPlantEventData is emitted by the place_mother_plant action.
// X, Y, Z are the tile coordinates where the plant should be spawned.
type PlaceMotherPlantEventData struct {
	X, Y, Z int
}

func (e PlaceMotherPlantEventData) GetType() event.EventType { return PlaceMotherPlant }

// ArmSelfDestructEventData is emitted when a player activates a self-destruct console.
type ArmSelfDestructEventData struct {
	Turns int // countdown length
}

func (e ArmSelfDestructEventData) GetType() event.EventType { return ArmSelfDestruct }

// LifePodEscapeEventData is emitted when a player activates a life pod console.
type LifePodEscapeEventData struct{}

func (e LifePodEscapeEventData) GetType() event.EventType { return LifePodEscape }

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
