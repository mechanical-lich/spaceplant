package component

import "github.com/mechanical-lich/mlge/ecs"

// DeadComponent .
type DeadComponent struct {
}

func (pc DeadComponent) GetType() ecs.ComponentType {
	return "DeadComponent"
}
