package component

// MyTurnComponent .
type StatsComponent struct {
	AC              int
	Str             int
	Dex             int
	Int             int
	Wis             int
	BasicAttackDice string
}

func (pc StatsComponent) GetType() string {
	return "StatsComponent"
}
