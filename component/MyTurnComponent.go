package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type MyTurnComponent struct {
}

func (pc MyTurnComponent) GetType() ecs.ComponentType {
	return "MyTurnComponent"
}
