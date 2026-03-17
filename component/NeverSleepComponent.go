package component

import "github.com/mechanical-lich/mlge/ecs"

// NocturnalComponent .
type NeverSleepComponent struct {
}

func (pc NeverSleepComponent) GetType() ecs.ComponentType {
	return "NeverSleepComponent"
}
