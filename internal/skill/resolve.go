package skill

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// SkillForAIType returns the first skill the entity has whose AIType matches,
// along with a ready-to-execute Action built from its first binding.
// Returns nil, nil if the entity has no matching skill or no action binding.
func SkillForAIType(entity *ecs.Entity, aiType string) (*SkillDef, action.Action) {
	if !entity.HasComponent(component.Skill) {
		return nil, nil
	}
	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	for _, skillID := range sc.Skills {
		def := Get(skillID)
		if def == nil || def.AIType != aiType {
			continue
		}
		for _, actionID := range def.ActionBindings {
			if act := action.CreateSkillAction(actionID, def.ActionParams); act != nil {
				return def, act
			}
		}
	}
	return nil, nil
}

// ActionForKey returns the action for the given action ID if one of the entity's
// skills provides a binding with that action ID, or nil if none match.
func ActionForKey(entity *ecs.Entity, actionID string) action.Action {
	if !entity.HasComponent(component.Skill) {
		return nil
	}

	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	for _, skillID := range sc.Skills {
		def := Get(skillID)
		if def == nil {
			continue
		}

		for _, boundActionID := range def.ActionBindings {
			if boundActionID == actionID {
				if act := action.CreateSkillAction(actionID, def.ActionParams); act != nil {
					return act
				}
			}
		}
	}
	return nil
}
