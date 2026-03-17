package component

import "github.com/mechanical-lich/mlge/ecs"

// HostileAIComponent .
type HostileAIComponent struct {
	SightRange int

	TargetX int
	TargetY int
}

func (pc HostileAIComponent) GetType() ecs.ComponentType {
	return "HostileAIComponent"
}
