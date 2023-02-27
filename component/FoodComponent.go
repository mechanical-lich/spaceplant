package component

// FoodComponent .
type FoodComponent struct {
	Amount int
}

func (pc FoodComponent) GetType() string {
	return "FoodComponent"
}
