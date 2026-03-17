package component

import "github.com/mechanical-lich/mlge/ecs"

// DescriptionComponent .
type DescriptionComponent struct {
	Name    string
	Faction string
}

func (pc DescriptionComponent) GetType() ecs.ComponentType {
	return "DescriptionComponent"
}
