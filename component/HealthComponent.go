package component

import "github.com/mechanical-lich/mlge/ecs"

// HealthComponent .
type HealthComponent struct {
	Health int
}

func (pc HealthComponent) GetType() ecs.ComponentType {
	return "HealthComponent"
}
