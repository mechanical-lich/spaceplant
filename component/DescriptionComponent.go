package component

// DescriptionComponent .
type DescriptionComponent struct {
	Name    string
	Faction string
}

func (pc DescriptionComponent) GetType() string {
	return "DescriptionComponent"
}
