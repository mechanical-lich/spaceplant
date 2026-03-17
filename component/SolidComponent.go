package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type SolidComponent struct {
}

func (pc SolidComponent) GetType() ecs.ComponentType {
	return "SolidComponent"
}
