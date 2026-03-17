package component

import "github.com/mechanical-lich/mlge/ecs"

type ArmorComponent struct {
	DefenseBonus int
}

func (pc ArmorComponent) GetType() ecs.ComponentType {
	return "ArmorComponent"
}

type WeaponComponent struct {
	AttackBonus int
	AttackDice  string
}

func (pc WeaponComponent) GetType() ecs.ComponentType {
	return "WeaponComponent"
}
