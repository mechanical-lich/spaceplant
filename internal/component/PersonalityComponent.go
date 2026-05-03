package component

import "github.com/mechanical-lich/mlge/ecs"

// PersonalityComponent stores traits that influence crew combat behaviour and reactions.
// Values range 0–100. These are mirrored into ScriptComponent.Vars at spawn time so the
// crew_ai.basic script can read them directly via get_var().
type PersonalityComponent struct {
	Bravery int // willingness to fight (low → flee, high → engage)
	Empathy int // likelihood to help allies and react strongly to bond deaths
	Loyalty int // obedience to player orders
	Sanity  int // drops on trauma; 0 → berserk
}

func (c *PersonalityComponent) GetType() ecs.ComponentType { return Personality }
