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
		applyMod(entity, mod)
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
		removeMod(entity, mod)
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
			applyMod(entity, mod)
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

// applyMod applies a single StatModifier to an entity.
func applyMod(entity *ecs.Entity, mod StatModifier) {
	if mod.Value != "" {
		applyStatValue(entity, mod.Stat, mod.Value)
	} else {
		applyStatDelta(entity, mod.Stat, mod.Delta)
	}
}

// removeMod reverses a single StatModifier on an entity.
func removeMod(entity *ecs.Entity, mod StatModifier) {
	if mod.Value != "" {
		removeStatValue(entity, mod.Stat, mod.Value)
	} else {
		applyStatDelta(entity, mod.Stat, -mod.Delta)
	}
}

// applyStatValue appends a string value to a slice-based stat field.
func applyStatValue(entity *ecs.Entity, stat, value string) {
	if !entity.HasComponent(component.Stats) {
		return
	}
	s := entity.GetComponent(component.Stats).(*component.StatsComponent)
	switch stat {
	case "resistance":
		if !slices.Contains(s.Resistances, value) {
			s.Resistances = append(s.Resistances, value)
		}
	}
}

// removeStatValue removes a string value from a slice-based stat field.
func removeStatValue(entity *ecs.Entity, stat, value string) {
	if !entity.HasComponent(component.Stats) {
		return
	}
	s := entity.GetComponent(component.Stats).(*component.StatsComponent)
	switch stat {
	case "resistance":
		if idx := slices.Index(s.Resistances, value); idx >= 0 {
			s.Resistances = slices.Delete(s.Resistances, idx, idx+1)
		}
	}
}

// applyStatDelta adds delta to a named numeric stat on the entity.
// Supported stat names for the AAG system:
//
//	speed         — EnergyComponent.Speed
//	ph            — StatsComponent.PH (Physique)
//	ag            — StatsComponent.AG (Agility)
//	ma            — StatsComponent.MA (Mental Ability)
//	cl            — StatsComponent.CL (Cool)
//	ld            — StatsComponent.LD (Leadership)
//	cs            — StatsComponent.CS (CombatSkill, ranged)
//	htcs          — StatsComponent.HTCS (Hand-to-Hand CombatSkill, melee)
//	natural_sp    — StatsComponent.NaturalSP (natural stopping power)
func applyStatDelta(entity *ecs.Entity, stat string, delta int) {
	switch stat {
	case "speed":
		if entity.HasComponent(component.Energy) {
			ec := entity.GetComponent(component.Energy).(*rlcomponents.EnergyComponent)
			ec.Speed += delta
		}
	case "ph":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).PH += delta
		}
	case "ag":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).AG += delta
		}
	case "ma":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).MA += delta
		}
	case "cl":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).CL += delta
		}
	case "ld":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).LD += delta
		}
	case "cs":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).CS += delta
		}
	case "htcs":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).HTCS += delta
		}
	case "natural_sp":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).NaturalSP += delta
		}
	case "natural_pen":
		if entity.HasComponent(component.Stats) {
			entity.GetComponent(component.Stats).(*component.StatsComponent).NaturalPen += delta
		}
	}
}
