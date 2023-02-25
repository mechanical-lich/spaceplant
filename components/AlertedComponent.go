package components

// MyTurnComponent .
type AlertedComponent struct {
	Duration int
}

func (pc AlertedComponent) GetType() string {
	return "AlertedComponent"
}

func (pc *AlertedComponent) Decay() bool {
	pc.Duration--

	return pc.Duration <= 0
}
