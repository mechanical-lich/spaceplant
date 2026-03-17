package component

import "github.com/mechanical-lich/mlge/ecs"

// WanderAIComponent .
type WanderAIComponent struct {
}

func (pc WanderAIComponent) GetType() ecs.ComponentType {
	return "WanderAIComponent"
}
