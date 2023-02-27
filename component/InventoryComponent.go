package component

// InventoryComponent .
type InventoryComponent struct {
	Bag []string
}

func (pc InventoryComponent) GetType() string {
	return "InventoryComponent"
}

func (iC *InventoryComponent) AddItem(item string) {
	iC.Bag = append(iC.Bag, item)
}
