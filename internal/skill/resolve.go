package skill

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// ActionForKey returns the action bound to a key by one of the entity's skills,
// or nil if no skill provides a binding for that key.
func ActionForKey(entity *ecs.Entity, key string) action.Action {
	if !entity.HasComponent(component.Skill) {
		return nil
	}

	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	for _, skillID := range sc.Skills {
		def := Get(skillID)
		if def == nil {
			continue
		}

		if actionID, ok := def.ActionBindings[key]; ok {
			if act := action.CreateSkillAction(actionID, def.ActionParams); act != nil {
				return act
			}
		}
	}
	return nil
}
