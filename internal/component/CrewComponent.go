package component

import "github.com/mechanical-lich/mlge/ecs"

// CrewComponent tracks a crew member's job, routine destinations, and current behavioural state.
type CrewComponent struct {
	Job   string // "engineer" | "medic" | "security" | "scientist" | "cook" | "officer"
	HomeX int    // tile x of their quarters (routine night destination)
	HomeY int
	HomeZ int
	WorkX int // tile x of their workstation (routine day destination)
	WorkY int
	WorkZ int
	State string // "routine" | "hunt" | "flee" | "berserk" | "follow" | "stay" | "ordered_attack"
}

func (c *CrewComponent) GetType() ecs.ComponentType { return CrewJob }
