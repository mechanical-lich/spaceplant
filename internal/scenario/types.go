package scenario

import "github.com/mechanical-lich/spaceplant/internal/wincondition"

// Scenario describes a self-contained game situation: which monsters spawn,
// how aggressively, which win/lose rules apply, and any bonus skills or
// backgrounds made available during character creation.
type Scenario struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Enabled          bool               `json:"enabled"`
	Hostiles         []string           `json:"hostiles"`
	RareHostiles     []string           `json:"rare_hostiles"`
	HostileInitial   int                `json:"hostile_initial"`
	HostileMax       int                `json:"hostile_max"`
	SpawnChance      float64            `json:"spawn_chance"` // 0.0–1.0
	ExtraSkills      []string           `json:"extra_skills"`
	ExtraBackgrounds []string           `json:"extra_backgrounds"`
	WinConditions    wincondition.RuleSet `json:"win_conditions"`
}
