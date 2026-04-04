package energy

// Base action costs. Movement is further modified by terrain via rlenergy.MoveCost.
const (
	CostMove   = 100
	CostAttack = 100
	CostQuick  = 50  // pickup, open door, heal, equip
	CostAimed  = 150 // aimed shot: deliberate aim, higher energy cost
	CostBurst  = 150 // burst fire: multiple rounds, higher energy cost
)
