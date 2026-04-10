package combat

import (
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
)

const CombatEventType mlgeevent.EventType = "CombatEvent"

// CombatEvent is queued on the event bus after each combat resolution.
// It carries enough information for the CombatListener to build player-facing messages.
type CombatEvent struct {
	X, Y, Z      int
	Attacker     *ecs.Entity
	Defender     *ecs.Entity
	AttackerName string
	DefenderName string
	Source       string // weapon name or action verb
	DamageType   string
	Damage       int
	BodyPart     string
	Miss         bool
	Broken       bool
	Amputated    bool
	Crit         bool
	Parried      bool
	SavePass     bool
	SaveFail     bool
	WoundPenalty int // CS penalty applied to this attack due to attacker wounds (> 0 means penalised)
}

func (e CombatEvent) GetType() mlgeevent.EventType { return CombatEventType }
