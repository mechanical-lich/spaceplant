package game

import (
	"math"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	spcombat "github.com/mechanical-lich/spaceplant/internal/combat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"

	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/game/listeners"
	"github.com/mechanical-lich/spaceplant/internal/system"
)

// compile-time assertion
var _ simulation.SimulationState = (*MainSimState)(nil)

// tickPhase tracks which stage of the turn cycle we are in.
type tickPhase int

const (
	// phaseWaitingForInput blocks until the player queues a command.
	phaseWaitingForInput tickPhase = iota

	// phaseRunningTick processes one tick: all entities with Energy > 0 act
	// once (player and NPCs unified). After acting, TickCount is incremented
	// and we check whether anyone still has energy for another tick.
	phaseRunningTick

	// phaseTurnComplete fires when no entity has energy left. It increments
	// TurnCount, resets TickCount, and calls AdvanceEnergy so every entity
	// gains Speed energy before the next tick.
	phaseTurnComplete
)

// MainSimState is the server-side gameplay state.
//
// Turn / Tick flow:
//
//	phaseTurnComplete  → AdvanceEnergy → phaseWaitingForInput (player can act)
//	                                   → phaseRunningTick     (only NPCs can act)
//	phaseWaitingForInput → player queues command → phaseRunningTick
//	phaseRunningTick → CleanUp → check turns → UpdateEntities → TickCount++
//	                 → player still has energy → phaseWaitingForInput
//	                 → only NPCs have energy   → phaseRunningTick
//	                 → nobody has energy       → phaseTurnComplete
type MainSimState struct {
	sim      *SimWorld
	done     bool
	phase    tickPhase
	npcDelay int
}

// NewMainSimState registers event listeners and returns a ready-to-use state.
// Starts at phaseTurnComplete so AdvanceEnergy runs before anyone acts.
func NewMainSimState(sim *SimWorld) *MainSimState {
	s := &MainSimState{
		sim:   sim,
		phase: phaseTurnComplete,
	}

	eventsystem.EventManager.RegisterListener(s, eventsystem.Stairs)
	eventsystem.EventManager.RegisterListener(s, eventsystem.DropItem)

	event.GetQueuedInstance().RegisterListener(
		&listeners.MessageListener{Sim: sim},
		message.MessageEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&listeners.InteractionListener{Sim: sim},
		rlcomponents.InteractionEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&listeners.PassoverListener{Sim: sim},
		rlcomponents.PassoverEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&listeners.CombatListener{Sim: sim},
		spcombat.CombatEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&listeners.ExcuseMeListener{Sim: sim},
		rlcomponents.ExcuseMeEventType,
	)

	event.GetQueuedInstance().RegisterListener(
		&listeners.DeathListener{Sim: sim},
		rlcomponents.DeathEventType,
	)

	return s
}

func (s *MainSimState) Done() bool { return s.done }

// ProcessCommand is called once per queued client command, before Tick.
// It pushes the key string onto the player's command queue.
func (s *MainSimState) ProcessCommand(cmd *transport.Command) {
	playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
	switch cmd.Type {
	case CmdAction:
		payload, ok := cmd.Payload.(ActionPayload)
		if !ok {
			return
		}
		playerC.PushCommand(payload.Key)
	case CmdReload:
		payload, ok := cmd.Payload.(ReloadPayload)
		if !ok {
			return
		}
		playerC.PendingReload = &component.PendingReloadData{
			WeaponItem: payload.WeaponItem,
			AmmoItem:   payload.AmmoItem,
		}
		// Push a sentinel so phaseWaitingForInput advances to phaseRunningTick.
		playerC.PushCommand("reload")
	case CmdAimedShot:
		payload, ok := cmd.Payload.(AimedShotPayload)
		if !ok {
			return
		}
		playerC.PendingAimedBodyPart = payload.BodyPart
		playerC.PushCommand("aimed_shot")
	}
}

// Tick advances the simulation by one server tick.
func (s *MainSimState) Tick(_ any) simulation.SimulationState {
	event.GetQueuedInstance().HandleQueue()

	if s.sim.Player == nil {
		return nil
	}

	switch s.phase {
	case phaseTurnComplete:
		s.sim.Mu.Lock()
		s.sim.TurnCount++
		s.sim.TickCount = 0
		playerGotTurn, _ := rlenergy.AdvanceEnergy(s.sim.Level.Entities, s.sim.Player)
		if playerGotTurn {
			s.phase = phaseWaitingForInput
		} else {
			s.phase = phaseRunningTick
		}
		s.sim.Mu.Unlock()

	case phaseWaitingForInput:
		playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
		if len(playerC.Commands) == 0 {
			return nil
		}
		s.phase = phaseRunningTick
		fallthrough

	case phaseRunningTick:
		if s.npcDelay > 0 {
			s.npcDelay--
			return nil
		}

		s.sim.Mu.Lock()

		// Resolve costs from the previous tick.
		system.CleanUpSystem{}.Update(s.sim.Level)

		// Re-grant turns to entities with leftover energy (multi-action).
		// No-op for entities that already hold MyTurn from AdvanceEnergy.
		rlenergy.RegrantTurns(s.sim.Level.Entities, s.sim.Player)

		// Check who currently holds MyTurn — covers both AdvanceEnergy-granted
		// and RegrantTurns-granted turns.
		playerGotTurn, anyGotTurn := s.anyHasMyTurn()

		if !anyGotTurn {
			s.phase = phaseTurnComplete
			s.sim.Mu.Unlock()
			return nil
		}

		// If the player has MyTurn but no command yet, the whole tick must
		// wait — remove MyTurn and park until input arrives.
		if playerGotTurn {
			playerC := s.sim.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
			if len(playerC.Commands) == 0 {
				s.sim.Player.RemoveComponent(rlcomponents.MyTurn)
				s.phase = phaseWaitingForInput
				s.sim.Mu.Unlock()
				return nil
			}
		}

		// Run all entity systems — player and NPCs in the same unified pass.
		s.sim.UpdateEntities()
		s.advanceAnimations()
		s.sim.TickCount++

		// Resolve the player's turn immediately so the client sees hasTurn=false
		// and stops sending commands. CanAct below then uses post-action energy.
		rlenergy.ResolveTurn(s.sim.Player)

		// Check NPC delay while TurnTaken is still on NPCs (before next CleanUp).
		if s.anyVisibleNPCActed() {
			s.npcDelay = config.Global().NpcTurnDelayTicks
		}

		if rlenergy.CanAct(s.sim.Player) {
			s.sim.Player.AddComponent(rlcomponents.GetMyTurn())
			s.phase = phaseWaitingForInput
		} else {
			s.phase = phaseRunningTick
		}

		s.sim.Mu.Unlock()
	}

	return nil
}

// anyHasMyTurn returns whether any entity currently holds MyTurn, and
// specifically whether the player does.
func (s *MainSimState) anyHasMyTurn() (playerHasTurn, anyHasTurn bool) {
	for _, e := range s.sim.Level.Entities {
		if e != nil && e.HasComponent(rlcomponents.MyTurn) {
			anyHasTurn = true
			if e == s.sim.Player {
				playerHasTurn = true
			}
		}
	}
	return
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
