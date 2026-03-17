package component

import "github.com/mechanical-lich/mlge/ecs"

// FoodComponent .
type FoodComponent struct {
	Amount int
}

func (pc FoodComponent) GetType() ecs.ComponentType {
	return "FoodComponent"
}
