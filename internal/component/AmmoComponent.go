package component

import "github.com/mechanical-lich/mlge/ecs"

// AmmoComponent marks an item as ammunition.
// AmmoType must match the WeaponComponent.AmmoType of the weapon it loads.
// Count is the number of rounds in this pack; the pack is removed from the
// inventory when Count reaches 0.
type AmmoComponent struct {
	AmmoType string
	Count    int
}

func (c *AmmoComponent) GetType() ecs.ComponentType {
	return Ammo
}
