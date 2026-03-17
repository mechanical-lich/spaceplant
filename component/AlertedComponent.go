package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type AlertedComponent struct {
	Duration int
}

func (pc AlertedComponent) GetType() ecs.ComponentType {
	return "AlertedComponent"
}

func (pc *AlertedComponent) Decay() bool {
	pc.Duration--

	return pc.Duration <= 0
}
