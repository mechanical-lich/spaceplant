package game

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/initiative"
	"github.com/mechanical-lich/spaceplant/internal/system"
)

// compile-time assertion
var _ simulation.SimulationState = (*MainSimState)(nil)

// MainSimState is the server-side gameplay state.
//
// Turn flow:
//   - waitingForPlayer == true  → block until the player queues a command, then run
//     CleanUp → UpdateEntities (player acts) → AdvanceInitiative.
//     If the player's counter hits zero again immediately, stay in waitingForPlayer;
//     otherwise enter the NPC phase with a configurable inter-turn delay.
//   - waitingForPlayer == false → NPC phase. Count down npcDelay; when it reaches
//     zero run CleanUp → AdvanceInitiative → UpdateEntities (NPCs act).
//     If player gets MyTurn, flip back to waitingForPlayer; else reset npcDelay.
type MainSimState struct {
	sim              *SimWorld
	done             bool
	waitingForPlayer bool
	npcDelay         int
}

// NewMainSimState registers event listeners and returns a ready-to-use state.
func NewMainSimState(sim *SimWorld) *MainSimState {
	s := &MainSimState{
		sim:              sim,
		waitingForPlayer: true,
	}

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
func (s *MainSimState) Tick(_ any) simulation.SimulationState {
	event.GetQueuedInstance().HandleQueue()

	if s.sim.Player == nil {
		return nil
	}

	playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)

	if s.waitingForPlayer {
		// Block until the player has queued a command.
		if len(playerC.Commands) == 0 {
			return nil
		}

		s.sim.Mu.Lock()
		system.CleanUpSystem{}.Update(s.sim.Level)
		s.sim.UpdateEntities() // player consumes their command and sets TurnTaken
		s.advanceAnimations()

		playerGotTurn, _ := initiative.AdvanceInitiative(s.sim.Level.Entities, s.sim.Player)
		if !playerGotTurn {
			// Enter NPC phase with no delay so NPCs react immediately to the
			// player's action. The delay is applied after each NPC round.
			s.waitingForPlayer = false
			s.npcDelay = 0
		}
		s.sim.Mu.Unlock()
	} else {
		// NPC phase: wait out the inter-turn delay, then let NPCs act.
		if s.npcDelay > 0 {
			s.npcDelay--
			return nil
		}
		s.sim.TurnCount++
		s.sim.Mu.Lock()
		system.CleanUpSystem{}.Update(s.sim.Level)
		playerGotTurn, anyGotTurn := initiative.AdvanceInitiative(s.sim.Level.Entities, s.sim.Player)

		// Strip the player's MyTurn so the PlayerSystem doesn't consume a
		// queued command during the NPC phase.
		if playerGotTurn {
			s.sim.Player.RemoveComponent(rlcomponents.MyTurn)
		}

		s.sim.UpdateEntities() // NPCs act; player is excluded
		s.advanceAnimations()

		if playerGotTurn {
			s.sim.Player.AddComponent(rlcomponents.GetMyTurn())
			s.waitingForPlayer = true
		} else if anyGotTurn {
			s.npcDelay = config.Global().NpcTurnDelayTicks
		}
		// If nothing fired, npcDelay stays 0 and we retry immediately next tick.
		s.sim.Mu.Unlock()
	}

	return nil
}

// advanceAnimations steps every entity's sprite animation cycle.
func (s *MainSimState) advanceAnimations() {
	for _, entity := range s.sim.Level.Entities {
		if entity.HasComponent("AppearanceComponent") {
			entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent).Update()
		}
	}
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
