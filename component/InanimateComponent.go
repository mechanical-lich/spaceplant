package component

import "github.com/mechanical-lich/mlge/ecs"

// InanimateComponent .
type InanimateComponent struct {
}

func (pc InanimateComponent) GetType() ecs.ComponentType {
	return "InanimateComponent"
}
