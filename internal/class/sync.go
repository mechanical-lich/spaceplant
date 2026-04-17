package class

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

// ApplyStatMods applies the StatMods from each of the entity's classes directly
// to its stats. Call this once after setting base stats during spawn.
func ApplyStatMods(entity *ecs.Entity) {
	if !entity.HasComponent(component.Class) {
		return
	}
	cc := entity.GetComponent(component.Class).(*component.ClassComponent)
	for _, classID := range cc.Classes {
		def := Get(classID)
		if def == nil {
			continue
		}
		for _, mod := range def.StatMods {
			skill.ApplyStatDelta(entity, mod.Stat, mod.Delta)
		}
	}
}

// SyncSkills applies all starting and chosen class skills to the entity. Call
// this after loading a save or spawning an entity that already has a class assigned.
func SyncSkills(entity *ecs.Entity) {
	if !entity.HasComponent(component.Class) {
		return
	}
	cc := entity.GetComponent(component.Class).(*component.ClassComponent)
	for _, classID := range cc.Classes {
		def := Get(classID)
		if def == nil {
			continue
		}
		for _, skillID := range def.StartingSkills {
			skill.Apply(entity, skillID)
		}
	}
	for _, skillID := range cc.ChosenSkills {
		skill.Apply(entity, skillID)
	}
}
