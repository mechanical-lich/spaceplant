package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type PoisonedComponent struct {
	Duration int
}

func (pc PoisonedComponent) GetType() ecs.ComponentType {
	return "PoisonedComponent"
}

func (pc *PoisonedComponent) Decay() bool {
	pc.Duration--

	return pc.Duration <= 0
}
