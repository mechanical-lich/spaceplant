package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type InteractComponent struct {
	Message []string
}

func (pc InteractComponent) GetType() ecs.ComponentType {
	return "InteractComponent"
}
