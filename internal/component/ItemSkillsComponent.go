package component

import "github.com/mechanical-lich/mlge/ecs"

// ItemSkillsComponent lists skills that an item grants to its wielder when equipped.
type ItemSkillsComponent struct {
	Skills []string
}

func (c *ItemSkillsComponent) GetType() ecs.ComponentType {
	return ItemSkills
}
