package skill

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mechanical-lich/spaceplant/internal/action"
)

// StatModifier adjusts a named stat.
// For numeric stats (speed, ac, str, dex, int, wis, melee_attack_bonus,
// ranged_attack_bonus) use Delta. For string-set stats (resistance, advantage)
// use Value; Delta is ignored.
type StatModifier struct {
	Stat  string `json:"stat"`
	Delta int    `json:"delta,omitempty"`
	Value string `json:"value,omitempty"`
}

// SkillDef is the JSON-loaded definition of a skill.
type SkillDef struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	StatMods    []StatModifier `json:"stat_modifiers,omitempty"`
	// ActionBindings adds new key → action mappings when this skill is active.
	// Keys are key-name strings (e.g. "K"); values are registered action IDs.
	ActionBindings map[string]string `json:"action_bindings,omitempty"`
	// ActionParams holds data-driven parameters passed to the action when executed.
	ActionParams action.ActionParams `json:"action_params,omitempty"`
	// AIType tells the AI system how to use this skill autonomously.
	// Supported values: "align_and_shoot" — get in range, line up, and fire.
	AIType string `json:"ai_type,omitempty"`
}

var registry = map[string]*SkillDef{}

// Load reads skill definitions from a JSON file (an array of SkillDef objects).
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("skill.Load: %w", err)
	}
	var defs []SkillDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return fmt.Errorf("skill.Load: %w", err)
	}
	for i := range defs {
		registry[defs[i].ID] = &defs[i]
	}
	return nil
}

// Get returns the SkillDef for the given ID, or nil if not found.
func Get(id string) *SkillDef {
	return registry[id]
}
