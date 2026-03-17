package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type StatsComponent struct {
	AC              int
	Str             int
	Dex             int
	Int             int
	Wis             int
	BasicAttackDice string
}

func (pc StatsComponent) GetType() ecs.ComponentType {
	return "StatsComponent"
}
