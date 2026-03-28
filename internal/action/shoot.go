package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ShootAction fires in a direction. Direction is a key name string
// (e.g. "W", "A", "S", "D").
type ShootAction struct {
	Direction string
}

func (a ShootAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostAttack
}

func (a ShootAction) Available(_ *ecs.Entity, _ *world.Level) bool {
	// TODO: check if entity has a ranged weapon
	return true
}

func (a ShootAction) Execute(entity *ecs.Entity, _ *world.Level) error {
	if a.Direction == "" {
		message.AddMessage("Wasn't given a direction to shoot!")
	} else {
		message.AddMessage("Shoot in the " + a.Direction + " direction!")
	}
	rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}
