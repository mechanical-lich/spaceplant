package component

// LightComponent .
type LightComponent struct {
	Brightness int
	Radius     int
	R, G, B    int
}

func (pc LightComponent) GetType() string {
	return "LightComponent"
}
