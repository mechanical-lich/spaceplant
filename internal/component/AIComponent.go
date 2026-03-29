package component

import "github.com/mechanical-lich/mlge/ecs"

// AdvancedAIComponent drives the state-machine AI system.
// Config fields are loaded from blueprints; runtime fields are managed by AdvancedAISystem.
type AdvancedAIComponent struct {
	// --- Config (set in blueprint) ---

	// Aggressiveness is a 0–100 chance per tick to enter the hunt state.
	// Proximity to the target increases this value further.
	Aggressiveness int `json:"Aggressiveness"`

	// Fear is a 0–100 chance per tick to enter the flee state.
	// When fear > aggressiveness the entity prefers flight over fight.
	Fear int `json:"Fear"`

	// SightRange is how many tiles away the entity can detect targets.
	SightRange int `json:"SightRange"`

	// Randomness adds 0–N jitter to each action score, preventing
	// deterministic behaviour.  Higher values produce more varied choices.
	Randomness int `json:"Randomness"`

	// AvoidsFriendlyFire causes the entity to skip area skills when a
	// friendly entity lies in the projected cone.
	AvoidsFriendlyFire bool `json:"AvoidsFriendlyFire"`

	// --- Runtime state (not loaded from blueprints) ---

	// State is the current FSM state: "idle", "hunt", or "flee".
	State string

	// TargetX/Y is the last known position of the hunt target.
	TargetX, TargetY int

	// FleeFromX/Y is the position being fled from.
	FleeFromX, FleeFromY int

	// LastAction is the name of the most recently executed action,
	// used to apply a repeat penalty when scoring.
	LastAction string

	// Path holds the most recent pathfinding result (tile indices).
	Path []int
}

func (c *AdvancedAIComponent) GetType() ecs.ComponentType {
	return AdvancedAI
}
