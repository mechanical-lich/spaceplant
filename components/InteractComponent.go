package components

// MyTurnComponent .
type InteractComponent struct {
	Message []string
}

func (pc InteractComponent) GetType() string {
	return "InteractComponent"
}
