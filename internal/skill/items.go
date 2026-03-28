package skill

import (
	"slices"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// SyncEquippedSkills reconciles the skills granted by equipped items with the
// entity's active skill list. Call this after any equip or unequip operation.
// Skills that are no longer provided by any equipped item are removed; skills
// newly provided by equipped items are applied.
func SyncEquippedSkills(entity *ecs.Entity) {
	desired := collectEquippedItemSkills(entity)

	sc := getOrAddSkillComponent(entity)

	// Remove skills no longer granted by any equipped item.
	for _, id := range sc.ItemSkills {
		if !slices.Contains(desired, id) {
			Remove(entity, id)
		}
	}

	// Build new ItemSkills list and apply newly granted skills.
	newItemSkills := make([]string, 0, len(desired))
	for _, id := range desired {
		Apply(entity, id)
		newItemSkills = append(newItemSkills, id)
	}
	sc.ItemSkills = newItemSkills
}

// collectEquippedItemSkills returns the deduplicated list of skill IDs granted
// by all items currently equipped on the entity.
func collectEquippedItemSkills(entity *ecs.Entity) []string {
	var skills []string

	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		for _, item := range []*ecs.Entity{
			inv.RightHand, inv.LeftHand,
			inv.Head, inv.Torso, inv.Legs, inv.Feet,
		} {
			skills = appendItemSkills(skills, item)
		}
	}

	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			skills = appendItemSkills(skills, item)
		}
	}

	return skills
}

func appendItemSkills(skills []string, item *ecs.Entity) []string {
	if item == nil || !item.HasComponent(component.ItemSkills) {
		return skills
	}
	isc := item.GetComponent(component.ItemSkills).(*component.ItemSkillsComponent)
	for _, id := range isc.Skills {
		if !slices.Contains(skills, id) {
			skills = append(skills, id)
		}
	}
	return skills
}
