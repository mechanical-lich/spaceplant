package class

import (
	"slices"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

func BuySkill(entity *ecs.Entity, skillID string) bool {
	if !entity.HasComponent(component.Class) {
		return false
	}

	c := entity.GetComponent(component.Class).(*component.ClassComponent)
	if c.UpgradePoints <= 0 {
		return false
	}
	if slices.Contains(c.ChosenSkills, skillID) {
		return false
	}

	for _, classID := range c.Classes {
		def := Get(classID)
		if def != nil && slices.Contains(def.Skills, skillID) {
			c.ChosenSkills = append(c.ChosenSkills, skillID)
			c.UpgradePoints--
			skill.Apply(entity, skillID)
			return true
		}
	}
	return false
}
