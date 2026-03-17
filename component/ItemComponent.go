package component

import "github.com/mechanical-lich/mlge/ecs"

type ItemComponent struct {
	Slot   string
	Effect string // TODO - heal, cure, buff, etc
	Value  int
}

func (ic ItemComponent) GetType() ecs.ComponentType {
	return "ItemComponent"
}
