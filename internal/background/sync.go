package background

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

// SyncSkills applies all skills from the entity's background.
// Call this after spawning an entity that has a BackgroundComponent.
func SyncSkills(entity *ecs.Entity) {
	if !entity.HasComponent(component.Background) {
		return
	}
	bc := entity.GetComponent(component.Background).(*component.BackgroundComponent)
	def := Get(bc.BackgroundID)
	if def == nil {
		return
	}
	for _, skillID := range def.Skills {
		skill.Apply(entity, skillID)
	}
}
