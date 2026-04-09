package component

import "github.com/mechanical-lich/mlge/ecs"

// RangeBand defines a CS modifier that applies up to MaxDist tiles away.
// Bands are evaluated in order; the first band where dist <= MaxDist wins.
// A band with MaxDist == 0 acts as a catch-all (applied at any distance).
type RangeBand struct {
	MaxDist int `json:"max_dist"`
	CSMod   int `json:"cs_mod"`
}

// WeaponComponent holds AAG/Phoenix Command combat stats for a weapon entity.
// It replaces rlcomponents.WeaponComponent; the component type key is kept
// identical ("WeaponComponent") so existing JSON blueprints load without change.
type WeaponComponent struct {
	// Penetration is the base damage value delivered on a successful hit.
	Penetration int `json:"Penetration"`
	// CombatSkillModifier is added to the wielder's CS roll (positive = easier to hit).
	CombatSkillModifier int `json:"CombatSkillModifier"`
	// DamageType is the elemental/physical damage category, e.g. "ballistic", "slashing".
	DamageType string `json:"DamageType"`
	// Ranged marks weapons that use the ranged attack path (ShootAction).
	Ranged bool `json:"Ranged"`
	// Range is the maximum effective range in tiles.
	Range int `json:"Range"`
	// RangeBands is an ordered list of distance thresholds to CS modifiers.
	// The first entry where dist <= MaxDist is used; if the list is empty the
	// global constants in shoot.go apply (point-blank -20, long-range -15).
	RangeBands []RangeBand `json:"RangeBands,omitempty"`
	// BurstSize is rounds per burst (0/1 = single shot only).
	BurstSize int `json:"BurstSize"`
	// SpreadAngle fires extra diagonal lines (0=single line, 1=3-wide, 2=5-wide).
	SpreadAngle int `json:"SpreadAngle"`
	// Ammo tracking.
	AmmoType    string `json:"AmmoType"`
	Magazine    int    `json:"Magazine"`
	MaxMagazine int    `json:"MaxMagazine"`
	// ActionCost overrides the default melee attack energy cost when set (> 0).
	// Lower values allow faster attacks (and thus more hits per turn at equal energy).
	ActionCost int `json:"ActionCost,omitempty"`
	// OnHitCondition is a status condition applied to the target on a successful hit
	// if they fail a resistance check. Supported values: "slowed", "poison", "burning".
	OnHitCondition string `json:"OnHitCondition,omitempty"`
	// OnHitDuration is the number of turns the on-hit condition lasts.
	OnHitDuration int `json:"OnHitDuration,omitempty"`
	// OnHitResistDC is the difficulty class for the target's resistance check.
	OnHitResistDC int `json:"OnHitResistDC,omitempty"`
	// OnHitCheckStat is the stat used for the resistance check (e.g. "ph", "cl").
	// Defaults to "cl" when unset.
	OnHitCheckStat string `json:"OnHitCheckStat,omitempty"`
	// Projectile sprite (used for visual FX, not combat).
	ProjectileX        int    `json:"ProjectileX"`
	ProjectileY        int    `json:"ProjectileY"`
	ProjectileResource string `json:"ProjectileResource"`
}

func (w WeaponComponent) GetType() ecs.ComponentType {
	return "WeaponComponent"
}

// RangeBandCSMod returns the CS modifier for the given distance using this
// weapon's RangeBands. Falls back to -1 sentinel if no bands are defined,
// signalling the caller to apply global defaults.
func (w *WeaponComponent) RangeBandCSMod(dist int) (int, bool) {
	if len(w.RangeBands) == 0 {
		return 0, false
	}
	for _, b := range w.RangeBands {
		if b.MaxDist == 0 || dist <= b.MaxDist {
			return b.CSMod, true
		}
	}
	// Beyond all defined bands — use last entry's modifier.
	return w.RangeBands[len(w.RangeBands)-1].CSMod, true
}
