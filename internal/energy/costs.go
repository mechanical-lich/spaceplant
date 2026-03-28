package energy

// Base action costs. Movement is further modified by terrain via rlenergy.MoveCost.
const (
	CostMove   = 100
	CostAttack = 100
	CostQuick  = 50 // pickup, open door, heal, equip
)
