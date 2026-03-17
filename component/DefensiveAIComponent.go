package component

import "github.com/mechanical-lich/mlge/ecs"

// DefensiveAIComponent .
type DefensiveAIComponent struct {
	AttackerX int
	AttackerY int
	Attacked  bool
}

func (pc DefensiveAIComponent) GetType() ecs.ComponentType {
	return "DefensiveAIComponent"
}
