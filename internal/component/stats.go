package component

import "github.com/mechanical-lich/mlge/ecs"

// StatsComponent holds Aliens Adventure Game (Phoenix Command) stats for an entity.
// Replaces the d20-based StatsComponent from ml-rogue-lib.
type StatsComponent struct {
	PH int // Physique: physical power, used for bare-hands penetration
	AG int // Agility: speed and coordination
	MA int // Mental Ability: intelligence
	CL int // Cool: composure under fire; drives CoolCheck resistance
	LD int // Leadership: command presence
	CS int // CombatSkill: percentile hit chance (1-100)

	NaturalSP  int // Natural stopping power (from skills like thick_skin)
	NaturalPen int // Natural penetration bonus added to bare-hands attacks (claws, vines, etc.)

	// Damage modifiers for resistances/weaknesses against damage types.
	Resistances []string // Half damage from these types
	Weaknesses  []string // Double damage from these types
}

func (s StatsComponent) GetType() ecs.ComponentType {
	return Stats
}
