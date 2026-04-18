package component

import "github.com/mechanical-lich/mlge/ecs"

const MotherPlantSeed ecs.ComponentType = "MotherPlantSeedComponent"

// MotherPlantSeedComponent tracks how many turns remain before the mobile
// mother plant transforms into the large rooted form.
type MotherPlantSeedComponent struct {
	TurnsLeft int
}

func (c *MotherPlantSeedComponent) GetType() ecs.ComponentType { return MotherPlantSeed }
