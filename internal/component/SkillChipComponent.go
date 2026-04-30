package component

import "github.com/mechanical-lich/mlge/ecs"

// SkillChipComponent marks an item as a skill chip that teaches the player a skill when consumed.
type SkillChipComponent struct {
	SkillId string
}

func (c *SkillChipComponent) GetType() ecs.ComponentType {
	return SkillChip
}
