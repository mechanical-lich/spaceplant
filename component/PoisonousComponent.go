package component

import "github.com/mechanical-lich/mlge/ecs"

// PoisonousComponent .
type PoisonousComponent struct {
	Duration int
}

func (pc PoisonousComponent) GetType() ecs.ComponentType {
	return "PoisonousComponent"
}
