package action

// ActionParams holds data-driven parameters for a skill's action.
// Fields default to zero values when not specified in JSON; actions should
// fall back to sensible defaults when a field is unset.
type ActionParams struct {
	// DamageDice is the damage roll expression, e.g. "2d6".
	DamageDice string `json:"damage_dice,omitempty"`
	// DamageType is the elemental or physical damage type, e.g. "fire".
	DamageType string `json:"damage_type,omitempty"`
	// SaveStat is the ability score used for the saving throw, e.g. "dex".
	SaveStat string `json:"save_stat,omitempty"`
	// ResistDC is the difficulty class for the resistance check the target must pass.
	ResistDC int `json:"resist_dc,omitempty"`
	// CheckStat is the stat used for the resistance check (e.g. "cl", "ph", "ag").
	// Defaults to "cl" (Cool) when unset.
	CheckStat string `json:"check_stat,omitempty"`
	// Range is the maximum reach in tiles for targeted or line effects.
	Range int `json:"range,omitempty"`
	// Radius is the radius in tiles for circular effects.
	Radius int `json:"radius,omitempty"`
	// Depth is the depth in tiles for cone effects.
	Depth int `json:"depth,omitempty"`
	// Spread is the perpendicular half-width for each row of a cone.
	// 0 or unset means only the centre tile (a line). -1 means the classic
	// widening cone where spread grows by 1 per depth row.
	Spread int `json:"spread,omitempty"`
	// Verb is the action word used in combat messages, e.g. "bite", "stab".
	Verb string `json:"verb,omitempty"`
	// ExtraDamageOnFailedSave is a dice expression for bonus damage dealt when
	// the target fails its saving throw, e.g. "3d6".
	ExtraDamageOnFailedSave string `json:"extra_damage_on_failed_save,omitempty"`
	// StatusConditionOnFailSave is the status condition applied to the target
	// on a failed save. Supported values: "poison", "burning".
	StatusConditionOnFailSave string `json:"status_condition_on_fail_save,omitempty"`
	// StatusConditionDuration is the number of turns the status condition lasts.
	StatusConditionDuration int `json:"status_condition_duration,omitempty"`
	// ActionCost overrides the default energy cost of the action.
	// When 0 (unset) the action uses its default cost (typically CostAttack = 100).
	ActionCost int `json:"action_cost,omitempty"`
}

// Cost returns ActionCost if set, otherwise the provided default.
func (p ActionParams) Cost(defaultCost int) int {
	if p.ActionCost > 0 {
		return p.ActionCost
	}
	return defaultCost
}
