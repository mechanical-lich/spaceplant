package component

import (
	"slices"

	"github.com/mechanical-lich/mlge/ecs"
)

// SkillComponent tracks the skill IDs active on an entity.
// ItemSkills is the subset of Skills that were granted by equipped items;
// it is managed by skill.SyncEquippedSkills and should not be edited directly.
type SkillComponent struct {
	Skills     []string
	ItemSkills []string
}

func (c *SkillComponent) GetType() ecs.ComponentType {
	return Skill
}

// HasSkill reports whether the given skill ID is active on this component.
func (c *SkillComponent) HasSkill(id string) bool {
	return slices.Contains(c.Skills, id)
}
