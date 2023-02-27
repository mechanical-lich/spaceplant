package component

// DamageComponent .
type DamageComponent struct {
	Amount int
}

func (pc DamageComponent) GetType() string {
	return "DamageComponent"
}
