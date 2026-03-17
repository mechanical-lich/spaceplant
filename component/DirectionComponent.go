package component

import "github.com/mechanical-lich/mlge/ecs"

// DirectionComponent .
type DirectionComponent struct {
	Direction int
}

func (pc DirectionComponent) GetType() ecs.ComponentType {
	return "DirectionComponent"
}
