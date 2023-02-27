package component

// HealthComponent .
type HealthComponent struct {
	Health int
}

func (pc HealthComponent) GetType() string {
	return "HealthComponent"
}
