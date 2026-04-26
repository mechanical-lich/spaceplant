package scenario

import "github.com/mechanical-lich/spaceplant/internal/wincondition"

// SpawnRule controls where a blueprint is allowed to spawn.
// Empty slices mean "anywhere" for that dimension.
type SpawnRule struct {
	Floors []string `json:"floors"` // floor theme names or indices ("0"–"5"); empty = any floor
	Rooms  []string `json:"rooms"`  // room tags; empty = anywhere on the allowed floors
}

// Scenario describes a self-contained game situation: which monsters spawn,
// how aggressively, which win/lose rules apply, and any bonus skills or
// backgrounds made available during character creation.
type Scenario struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	Enabled          bool                 `json:"enabled"`
	Hostiles         []string             `json:"hostiles"`
	RareHostiles     []string             `json:"rare_hostiles"`
	HostileInitial   int                  `json:"hostile_initial"`
	HostileMax       int                  `json:"hostile_max"`
	SpawnChance      float64              `json:"spawn_chance"` // 0.0–1.0
	ExtraSkills      []string             `json:"extra_skills"`
	ExtraBackgrounds []string             `json:"extra_backgrounds"`
	BossSpawns       []string             `json:"boss_spawns"`    // blueprint IDs spawned once at station creation
	SpawnRules       map[string]SpawnRule `json:"spawn_rules"`    // per-blueprint placement rules
	SetupScripts     []string             `json:"setup_scripts"`  // script paths run once after generation, in order
	WinConditions    wincondition.RuleSet `json:"win_conditions"`
}
