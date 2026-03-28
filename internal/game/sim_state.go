package game

import (
	"math"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"

	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/system"
)

// compile-time assertion
var _ simulation.SimulationState = (*MainSimState)(nil)

// MainSimState is the server-side gameplay state.
//
// Turn flow (energy-based):
//   - waitingForPlayer == true  → block until the player queues a command, then run
//     CleanUp → re-grant MyTurn if energy still at threshold → UpdateEntities.
//     If the player still has enough energy after cost deduction, they act again
//     immediately (no game time passes). Otherwise enter the NPC phase.
//   - waitingForPlayer == false → NPC phase. CleanUp deducts costs; any NPC with
//     leftover energy >= threshold gets MyTurn again (multi-action, no time advance).
//     When no one has leftover energy, AdvanceEnergy ticks everyone. Visible NPC
//     actions incur a configurable inter-turn delay for visual feedback.
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

		// Process the player's action first, before any other entity.
		s.sim.UpdatePlayer()

		// Immediately deduct cost if the player consumed their turn.
		// This avoids a one-tick delay between acting and cost resolution.
		rlenergy.ResolveTurn(s.sim.Player)

		// Full world update — lighting, doors, status effects, etc.
		// Player won't re-act here because MyTurn was already stripped.
		s.sim.UpdateEntities()
		s.advanceAnimations()

		// Decide what happens next.
		if rlenergy.CanAct(s.sim.Player) {
			// Player has enough energy to act again (multi-action).
			s.sim.Player.AddComponent(rlcomponents.GetMyTurn())
		} else {
			// Advance time so the player doesn't wait an extra tick.
			playerGotTurn, _ := rlenergy.AdvanceEnergy(s.sim.Level.Entities, s.sim.Player)
			if !playerGotTurn {
				s.waitingForPlayer = false
				s.npcDelay = 0
			}
		}
		s.sim.Mu.Unlock()
	} else {
		// NPC phase: wait out the inter-turn delay, then let NPCs act.
		if s.npcDelay > 0 {
			s.npcDelay--
			return nil
		}

		s.sim.Mu.Lock()
		system.CleanUpSystem{}.Update(s.sim.Level)

		// Re-grant turns to entities that still have leftover energy (multi-action).
		playerGotTurn, anyGotTurn := rlenergy.RegrantTurns(s.sim.Level.Entities, s.sim.Player)

		if !anyGotTurn {
			// No leftover energy anywhere — advance time (tick everyone).
			s.sim.TurnCount++
			playerGotTurn, anyGotTurn = rlenergy.AdvanceEnergy(s.sim.Level.Entities, s.sim.Player)
		}

		// Strip the player's MyTurn so the PlayerSystem doesn't consume a
		// queued command during the NPC phase.
		if playerGotTurn {
			s.sim.Player.RemoveComponent(rlcomponents.MyTurn)
		}

		if anyGotTurn {
			s.sim.UpdateEntities()
			s.advanceAnimations()
		}

		if playerGotTurn {
			s.sim.Player.AddComponent(rlcomponents.GetMyTurn())
			s.waitingForPlayer = true
		} else if anyGotTurn {
			// Only delay if a visible NPC acted — off-screen actions resolve instantly.
			if s.anyVisibleNPCActed() {
				s.npcDelay = config.Global().NpcTurnDelayTicks
			}
		}
		// If nothing fired, npcDelay stays 0 and we retry immediately next tick.
		s.sim.Mu.Unlock()
	}

	return nil
}

// anyVisibleNPCActed returns true if any non-player entity that just acted
// (has TurnTaken) is within the viewport around the player.
func (s *MainSimState) anyVisibleNPCActed() bool {
	if s.sim.Player == nil || !s.sim.Player.HasComponent(component.Position) {
		return false
	}
	ppc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := ppc.GetX(), ppc.GetY(), ppc.GetZ()

	cfg := config.Global()
	scale := math.Round(cfg.RenderScale)
	if scale < 1 {
		scale = 1
	}
	halfW := int(math.Ceil(float64(cfg.WorldWidth)/(float64(cfg.SpriteSizeW)*scale))) / 2
	halfH := int(math.Ceil(float64(cfg.WorldHeight)/(float64(cfg.SpriteSizeH)*scale))) / 2

	for _, entity := range s.sim.Level.Entities {
		if entity == s.sim.Player {
			continue
		}
		if !entity.HasComponent(rlcomponents.TurnTaken) || !entity.HasComponent(component.Position) {
			continue
		}
		epc := entity.GetComponent(component.Position).(*component.PositionComponent)
		ex, ey, ez := epc.GetX(), epc.GetY(), epc.GetZ()
		if ez != pz {
			continue
		}
		dx := ex - px
		dy := ey - py
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= halfW && dy <= halfH {
			return true
		}
	}
	return false
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
