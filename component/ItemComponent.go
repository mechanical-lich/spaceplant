package component

type ItemComponent struct {
	Slot   string
	Effect string // TODO - heal, cure, buff, etc
	Value  int
}

func (ic ItemComponent) GetType() string {
	return "ItemComponent"
}
