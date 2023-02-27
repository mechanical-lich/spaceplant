package component

// ShopComponent .
type ShopComponent struct {
	ItemsForSale []string
}

func (pc ShopComponent) GetType() string {
	return "ShopComponent"
}
