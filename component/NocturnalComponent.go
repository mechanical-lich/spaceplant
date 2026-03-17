package component

import "github.com/mechanical-lich/mlge/ecs"

// NocturnalComponent .
type NocturnalComponent struct {
}

func (pc NocturnalComponent) GetType() ecs.ComponentType {
	return "NocturnalComponent"
}
