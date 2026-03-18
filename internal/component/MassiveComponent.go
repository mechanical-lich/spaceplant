package component

import "github.com/mechanical-lich/mlge/ecs"

// MassiveComponent .
type MassiveComponent struct {
}

func (pc MassiveComponent) GetType() ecs.ComponentType {
	return "MassiveComponent"
}
