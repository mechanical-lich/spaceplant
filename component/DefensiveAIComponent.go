package component

// DefensiveAIComponent .
type DefensiveAIComponent struct {
	AttackerX int
	AttackerY int
	Attacked  bool
}

func (pc DefensiveAIComponent) GetType() string {
	return "DefensiveAIComponent"
}
