package components

// MyTurnComponent .
type AppearanceComponent struct {
	SpriteX  int
	SpriteY  int
	Resource string
	R, G, B  uint8
}

func (pc AppearanceComponent) GetType() string {
	return "AppearanceComponent"
}
