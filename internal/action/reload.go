package action

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ReloadAction loads ammo from AmmoItem into WeaponItem's magazine.
// Both entities must still be in the actor's inventory at execution time.
type ReloadAction struct {
	WeaponItem *ecs.Entity
	AmmoItem   *ecs.Entity
}

func (a ReloadAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostQuick
}

func (a ReloadAction) Available(_ *ecs.Entity, _ *world.Level) bool {
	return a.WeaponItem != nil && a.AmmoItem != nil
}

func (a ReloadAction) Execute(entity *ecs.Entity, _ *world.Level) error {
	if a.WeaponItem == nil || a.AmmoItem == nil {
		return nil
	}
	if !a.WeaponItem.HasComponent(component.Weapon) || !a.AmmoItem.HasComponent(component.Ammo) {
		return nil
	}

	wc := a.WeaponItem.GetComponent(component.Weapon).(*component.WeaponComponent)
	ac := a.AmmoItem.GetComponent(component.Ammo).(*component.AmmoComponent)

	if ac.AmmoType != wc.AmmoType {
		message.AddMessage("Incompatible ammo type.")
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	needed := wc.MaxMagazine - wc.Magazine
	if needed <= 0 {
		message.AddMessage("Magazine already full.")
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	loaded := needed
	if ac.Count < loaded {
		loaded = ac.Count
	}

	wc.Magazine += loaded
	ac.Count -= loaded

	if ac.Count <= 0 {
		removeFromInventory(entity, a.AmmoItem)
	}

	ic := a.WeaponItem.GetComponent(component.Item).(*component.ItemComponent)
	message.AddMessage(fmt.Sprintf("Loaded %d rounds into %s.", loaded, ic.Name))
	rlenergy.SetActionCost(entity, energy.CostQuick)
	return nil
}
