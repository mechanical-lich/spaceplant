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
	// SaveDC is the difficulty class for the saving throw.
	SaveDC int `json:"save_dc,omitempty"`
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
}
