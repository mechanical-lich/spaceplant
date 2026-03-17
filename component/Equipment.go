package component

type ArmorComponent struct {
	DefenseBonus int
}

func (pc ArmorComponent) GetType() string {
	return "ArmorComponent"
}

type WeaponComponent struct {
	AttackBonus int
	AttackDice  string
}

func (pc WeaponComponent) GetType() string {
	return "WeaponComponent"
}
