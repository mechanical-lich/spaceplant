package component

import "github.com/mechanical-lich/mlge/ecs"

// InitiativeComponent .
type InitiativeComponent struct {
	DefaultValue  int
	OverrideValue int
	Ticks         int
}

func (pc InitiativeComponent) GetType() ecs.ComponentType {
	return "InitiativeComponent"
}
