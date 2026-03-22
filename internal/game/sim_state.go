package game

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/system"
)

// compile-time assertion
var _ simulation.SimulationState = (*MainSimState)(nil)

// MainSimState is the server-side gameplay state. It owns the turn-based gate:
// the simulation only advances when the player has queued an action command.
type MainSimState struct {
	sim  *SimWorld
	done bool
}

// NewMainSimState registers event listeners and returns a ready-to-use state.
func NewMainSimState(sim *SimWorld) *MainSimState {
	s := &MainSimState{sim: sim}

	eventsystem.EventManager.RegisterListener(s, eventsystem.Stairs)
	eventsystem.EventManager.RegisterListener(s, eventsystem.DropItem)

	event.GetQueuedInstance().RegisterListener(
		&queuedMessageListener{level: sim.Level, player: sim.Player},
		message.MessageEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&interactionListener{sim: sim},
		rlcomponents.InteractionEventType,
	)

	return s
}

func (s *MainSimState) Done() bool { return s.done }

// ProcessCommand is called once per queued client command, before Tick.
// It pushes the key string onto the player's command queue.
func (s *MainSimState) ProcessCommand(cmd *transport.Command) {
	if cmd.Type != CmdAction {
		return
	}
	payload, ok := cmd.Payload.(ActionPayload)
	if !ok {
		return
	}
	playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
	playerC.PushCommand(payload.Key)
}

// Tick advances the simulation by one server tick.
// The turn-based gate: advance only when it is not the player's turn, or when
// the player has a command queued (meaning they have acted this tick).
func (s *MainSimState) Tick(_ any) simulation.SimulationState {
	event.GetQueuedInstance().HandleQueue()

	if s.sim.Player == nil {
		return nil
	}

	playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
	playerHasTurn := s.sim.Player.HasComponent("MyTurn")
	shouldAdvance := !playerHasTurn || len(playerC.Commands) > 0

	if shouldAdvance {
		s.sim.Mu.Lock()
		system.CleanUpSystem{}.Update(s.sim.Level)
		s.sim.UpdateEntities()
		s.sim.Mu.Unlock()
	}

	return nil
}

// HandleEvent responds to Stairs and DropItem events fired by systems.
func (s *MainSimState) HandleEvent(data event.EventData) error {
	switch data.GetType() {
	case eventsystem.Stairs:
		stairsEvent := data.(eventsystem.StairsEventData)
		pc := s.sim.Player.GetComponent("Position").(*component.PositionComponent)

		if stairsEvent.Up {
			if s.sim.CurrentZ < numLevels-1 {
				s.sim.CurrentZ++
				s.sim.Level.PlaceEntity(pc.GetX(), pc.GetY(), s.sim.CurrentZ, s.sim.Player)
			}
		} else {
			if s.sim.CurrentZ > 0 {
				s.sim.CurrentZ--
				s.sim.Level.PlaceEntity(pc.GetX(), pc.GetY(), s.sim.CurrentZ, s.sim.Player)
			}
		}
		s.sim.UpdateEntities()

	case eventsystem.DropItem:
		dropItemEvent := data.(eventsystem.DropItemEventData)
		dropItemEvent.Item.GetComponent("Position").(*component.PositionComponent).
			SetPosition(dropItemEvent.X, dropItemEvent.Y, dropItemEvent.Z)
		s.sim.Level.AddEntity(dropItemEvent.Item)
	}
	return nil
}
