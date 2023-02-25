package components

// AttackComponent .
type AttackComponent struct {
	X       int
	Y       int
	Frame   int
	SpriteX int
	SpriteY int
	CleanUp bool
}

func (pc AttackComponent) GetType() string {
	return "AttackComponent"
}
