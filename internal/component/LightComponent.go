package component

import "github.com/mechanical-lich/mlge/ecs"

// LightComponent .
type LightComponent struct {
	Brightness int
	Radius     int
	R, G, B    int
}

func (pc LightComponent) GetType() ecs.ComponentType {
	return "LightComponent"
}
