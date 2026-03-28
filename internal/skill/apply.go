package skill

import (
	"slices"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// Apply adds a skill to an entity and immediately applies its stat modifiers.
// Does nothing if the skill is already active or the ID is unknown.
func Apply(entity *ecs.Entity, skillID string) {
	def := Get(skillID)
	if def == nil {
		return
	}
	sc := getOrAddSkillComponent(entity)
	if slices.Contains(sc.Skills, skillID) {
		return
	}
	sc.Skills = append(sc.Skills, skillID)
	for _, mod := range def.StatMods {
		applyStatDelta(entity, mod.Stat, mod.Delta)
	}
}

// Remove removes a skill from an entity, reverting its stat modifiers.
// Does nothing if the skill is not active or the ID is unknown.
func Remove(entity *ecs.Entity, skillID string) {
	def := Get(skillID)
	if def == nil || !entity.HasComponent(component.Skill) {
		return
	}
	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	idx := slices.Index(sc.Skills, skillID)
	if idx < 0 {
		return
	}
	sc.Skills = slices.Delete(sc.Skills, idx, idx+1)
	for _, mod := range def.StatMods {
		applyStatDelta(entity, mod.Stat, -mod.Delta)
	}
}

// Initialize applies stat modifiers for all skills already listed in the entity's
// SkillComponent. Call this once after creating an entity from a blueprint so
// that JSON-defined skills take effect. It is safe to call on entities that have
// no SkillComponent (no-op).
func Initialize(entity *ecs.Entity) {
	if !entity.HasComponent(component.Skill) {
		return
	}
	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	for _, skillID := range sc.Skills {
		def := Get(skillID)
		if def == nil {
			continue
		}
		for _, mod := range def.StatMods {
			applyStatDelta(entity, mod.Stat, mod.Delta)
		}
	}
}

// HasSkill reports whether the entity currently has the given skill active.
func HasSkill(entity *ecs.Entity, skillID string) bool {
	if !entity.HasComponent(component.Skill) {
		return false
	}
	sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
	return slices.Contains(sc.Skills, skillID)
}

// getOrAddSkillComponent returns the entity's SkillComponent, creating it if absent.
func getOrAddSkillComponent(entity *ecs.Entity) *component.SkillComponent {
	if entity.HasComponent(component.Skill) {
		return entity.GetComponent(component.Skill).(*component.SkillComponent)
	}
	sc := &component.SkillComponent{}
	entity.AddComponent(sc)
	return sc
}

// applyStatDelta adds delta to a named stat on the entity.
// Supported names: "speed", "ac", "str", "dex", "int", "wis".
func applyStatDelta(entity *ecs.Entity, stat string, delta int) {
	switch stat {
	case "speed":
		if entity.HasComponent(component.Energy) {
			ec := entity.GetComponent(component.Energy).(*rlcomponents.EnergyComponent)
			ec.Speed += delta
		}
	case "ac":
		if entity.HasComponent(component.Stats) {
			s := entity.GetComponent(component.Stats).(*rlcomponents.StatsComponent)
			s.AC += delta
		}
	case "str":
		if entity.HasComponent(component.Stats) {
			s := entity.GetComponent(component.Stats).(*rlcomponents.StatsComponent)
			s.Str += delta
		}
	case "dex":
		if entity.HasComponent(component.Stats) {
			s := entity.GetComponent(component.Stats).(*rlcomponents.StatsComponent)
			s.Dex += delta
		}
	case "int":
		if entity.HasComponent(component.Stats) {
			s := entity.GetComponent(component.Stats).(*rlcomponents.StatsComponent)
			s.Int += delta
		}
	case "wis":
		if entity.HasComponent(component.Stats) {
			s := entity.GetComponent(component.Stats).(*rlcomponents.StatsComponent)
			s.Wis += delta
		}
	}
}
