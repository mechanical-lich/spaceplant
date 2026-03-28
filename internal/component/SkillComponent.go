package component

import "github.com/mechanical-lich/mlge/ecs"

// SkillComponent tracks the skill IDs active on an entity.
type SkillComponent struct {
	Skills []string
}

func (c *SkillComponent) GetType() ecs.ComponentType {
	return Skill
}
