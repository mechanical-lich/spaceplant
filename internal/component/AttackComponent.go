package component

import "github.com/mechanical-lich/mlge/ecs"

// AttackComponent .
type AttackComponent struct {
	X       int
	Y       int
	Frame   int
	SpriteX int
	SpriteY int
	CleanUp bool
}

func (pc AttackComponent) GetType() ecs.ComponentType {
	return "AttackComponent"
}
