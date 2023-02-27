package component

// HostileAIComponent .
type HostileAIComponent struct {
	SightRange int

	TargetX int
	TargetY int
}

func (pc HostileAIComponent) GetType() string {
	return "HostileAIComponent"
}
