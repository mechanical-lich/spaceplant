package combat

import (
	"math/rand"
	"slices"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const bareHandsPenBase = 3 // base Pen for unarmed attack before PH modifier

// Hit resolves a melee attack from attacker against defender using the
// Aliens Adventure Game (Phoenix Command) combat system.
// Returns true if the attack landed.
func Hit(level *world.Level, attacker, defender *ecs.Entity) bool {
	return hitCore(level, attacker, defender, nil, -1, 0, "")
}

// HitWithPen resolves an attack with an optional Pen override.
// If penOverride >= 0 it replaces the weapon's Penetration value (used by
// special unarmed attacks like hands_only).
func HitWithPen(level *world.Level, attacker, defender *ecs.Entity, penOverride int) bool {
	return hitCore(level, attacker, defender, nil, penOverride, 0, "")
}

// HitRanged resolves a ranged attack with a specific weapon and CS bonus/penalty.
// weaponOverride pins which weapon is used for CS modifier, Pen, and damage type,
// preventing non-deterministic map iteration from selecting the wrong equipped item.
// csBonus is added on top of the weapon's CombatSkillModifier.
func HitRanged(level *world.Level, attacker, defender *ecs.Entity, weaponOverride *component.WeaponComponent, csBonus int) bool {
	return hitCore(level, attacker, defender, weaponOverride, -1, csBonus, "")
}

// HitRangedTargeted resolves a ranged attack biased toward a specific body part.
// On hit, the chosen body part has a CS-scaled chance (~75% at CS 50) of being struck.
func HitRangedTargeted(level *world.Level, attacker, defender *ecs.Entity, weaponOverride *component.WeaponComponent, csBonus int, aimedBodyPart string) bool {
	return hitCore(level, attacker, defender, weaponOverride, -1, csBonus, aimedBodyPart)
}

// hitCore is the shared hit-resolution engine.
// weaponOverride, when non-nil, is used in place of equippedWeapon(attacker).
// penOverride < 0 means use the weapon / bare-hands Pen. csBonus is added to CS
// after the weapon's CombatSkillModifier (range bands, aimed shot, etc.).
func hitCore(level *world.Level, attacker, defender *ecs.Entity, weaponOverride *component.WeaponComponent, penOverride, csBonus int, aimedBodyPart string) bool {
	apc := attacker.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	dpc := defender.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

	attackerName := rlentity.GetName(attacker)
	defenderName := rlentity.GetName(defender)

	// --- Determine effective CombatSkill ---
	// Melee (weaponOverride == nil) uses HTCS + PH bracket bonus.
	// Ranged (weaponOverride != nil) uses CS.
	cs := 30 // fallback if no stats
	if attacker.HasComponent(component.Stats) {
		sc := attacker.GetComponent(component.Stats).(*component.StatsComponent)
		if weaponOverride == nil {
			cs = sc.HTCS + statMeleeBonus(sc.PH)
		} else {
			cs = sc.CS
		}
	}

	weapon := weaponOverride
	if weapon == nil {
		weapon = equippedWeapon(attacker)
	}
	if weapon != nil {
		cs += weapon.CombatSkillModifier
	}
	cs += csBonus
	if cs < 1 {
		cs = 1
	}

	// --- Roll to hit (d100 <= CS) ---
	roll, _ := dice.Roll("1d100")
	miss := roll > cs

	weaponName := "fist"
	damageType := "bludgeoning"
	if weapon != nil && weapon.DamageType != "" {
		damageType = weapon.DamageType
		if weaponOverride == nil {
			if name := equippedWeaponItemName(attacker); name != "" {
				weaponName = name
			} else {
				weaponName = damageType
			}
		} else {
			weaponName = damageType
		}
	}

	if miss {
		mlgeevent.GetQueuedInstance().QueueEvent(CombatEvent{
			X: dpc.GetX(), Y: dpc.GetY(), Z: dpc.GetZ(),
			Attacker:     attacker,
			Defender:     defender,
			AttackerName: attackerName,
			DefenderName: defenderName,
			Source:       weaponName,
			DamageType:   damageType,
			Miss:         true,
		})
		return false
	}

	// --- Determine hit body part ---
	partName := pickBodyPart(defender, attacker, aimedBodyPart)

	// --- Determine Penetration ---
	pen := penOverride
	if pen < 0 {
		if weapon != nil {
			pen = weapon.Penetration
		} else {
			// Bare hands: base + PH bonus + NaturalPen
			ph := 10
			naturalPen := 0
			if attacker.HasComponent(component.Stats) {
				sc := attacker.GetComponent(component.Stats).(*component.StatsComponent)
				ph = sc.PH
				naturalPen = sc.NaturalPen
			}
			pen = bareHandsPenBase + ph/4 + naturalPen
		}
	}

	// --- Determine Stopping Power for hit part ---
	sp := armorSP(defender, partName)

	// --- Apply damage ---
	damage := pen - sp
	if damage < 0 {
		damage = 0
	}

	// --- Parry roll (melee only) ---
	parried := false
	if weaponOverride == nil && damage > 0 {
		htcs, ag := 20, 10
		if defender.HasComponent(component.Stats) {
			sc := defender.GetComponent(component.Stats).(*component.StatsComponent)
			htcs = sc.HTCS
			ag = sc.AG
		}
		parryThreshold := htcs + statMeleeBonus(ag)
		if parryThreshold < 1 {
			parryThreshold = 1
		}
		parryRoll, _ := dice.Roll("1d100")
		if parryRoll <= parryThreshold {
			damage /= 2
			parried = true
		}
	}

	broken, amputated, killed := false, false, false
	if damage > 0 && partName != "" && defender.HasComponent(rlcomponents.Body) {
		bc := defender.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
		broken, amputated, killed = applyPartDamage(bc, partName, damage, damageType, defender)
	} else if damage > 0 && defender.HasComponent(rlcomponents.Health) {
		hc := defender.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
		hc.Health -= damage
		if hc.Health <= 0 {
			hc.Health = 0
			killed = true
		}
	}

	_ = apc
	if killed {
		defender.AddComponent(&rlcomponents.DeadComponent{})
	}

	mlgeevent.GetQueuedInstance().QueueEvent(CombatEvent{
		X: dpc.GetX(), Y: dpc.GetY(), Z: dpc.GetZ(),
		Attacker:     attacker,
		Defender:     defender,
		AttackerName: attackerName,
		DefenderName: defenderName,
		Source:       weaponName,
		DamageType:   damageType,
		Damage:       damage,
		BodyPart:     partName,
		Parried:      parried,
	})

	if broken || amputated {
		mlgeevent.GetQueuedInstance().QueueEvent(CombatEvent{
			X: dpc.GetX(), Y: dpc.GetY(), Z: dpc.GetZ(),
			Attacker:     attacker,
			Defender:     defender,
			AttackerName: attackerName,
			DefenderName: defenderName,
			BodyPart:     partName,
			Broken:       broken,
			Amputated:    amputated,
		})
	}

	return true
}

// CoolCheck resolves a Cool-based resistance roll against a difficulty value.
// Returns true if the entity resists (succeeds). dc is added difficulty (higher = harder).
// Threshold = CL * 5; success if d100 <= threshold - dc.
func CoolCheck(entity *ecs.Entity, dc int) bool {
	cl := 10
	if entity.HasComponent(component.Stats) {
		cl = entity.GetComponent(component.Stats).(*component.StatsComponent).CL
	}
	threshold := cl*5 - dc
	if threshold < 1 {
		threshold = 1
	}
	roll, _ := dice.Roll("1d100")
	return roll <= threshold
}

// statMeleeBonus returns a CS modifier from a stat value using Phoenix Command brackets.
// Used for both PH (melee offense) and AG (melee defense/parry).
func statMeleeBonus(stat int) int {
	switch {
	case stat >= 18:
		return 20
	case stat >= 16:
		return 15
	case stat >= 14:
		return 10
	case stat >= 12:
		return 5
	case stat >= 10:
		return 0
	default:
		return -10
	}
}

// --- helpers ---

// equippedWeaponItemName returns the display name of the first equipped melee weapon item, or "".
// Ranged weapons are excluded to match equippedWeapon's selection logic.
func equippedWeaponItemName(entity *ecs.Entity) string {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return rlentity.GetName(item)
				}
			}
		}
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
		for _, item := range []*ecs.Entity{inv.RightHand, inv.LeftHand} {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return rlentity.GetName(item)
				}
			}
		}
	}
	return ""
}

// equippedWeapon returns the first equipped melee WeaponComponent found on the entity, or nil.
// Ranged weapons are excluded to prevent melee attacks from inheriting ranged weapon stats.
func equippedWeapon(entity *ecs.Entity) *component.WeaponComponent {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return wc
				}
			}
		}
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
		for _, item := range []*ecs.Entity{inv.RightHand, inv.LeftHand} {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return wc
				}
			}
		}
	}
	return nil
}

// pickBodyPart selects a body part name using HitLocationComponent weights.
// Falls back to equal-weight selection if HitLocationComponent is absent.
// If aimedBodyPart is non-empty, the attacker's CS gives a biased chance
// (~75% at CS 50, clamped 60–90%) of hitting the chosen part directly.
func pickBodyPart(entity *ecs.Entity, attacker *ecs.Entity, aimedBodyPart string) string {
	if !entity.HasComponent(rlcomponents.Body) {
		return ""
	}
	bc := entity.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)

	// If a body part was aimed at and it isn't amputated, apply CS-based bias.
	if aimedBodyPart != "" {
		if part, ok := bc.Parts[aimedBodyPart]; ok && !part.Amputated {
			cs := 50
			if attacker != nil && attacker.HasComponent(component.Stats) {
				cs = attacker.GetComponent(component.Stats).(*component.StatsComponent).CS
			}
			// chance = 75 + (CS-50)/10, clamped to [60, 90]
			chance := 75 + (cs-50)/10
			if chance < 60 {
				chance = 60
			}
			if chance > 90 {
				chance = 90
			}
			if rand.Intn(100) < chance {
				return aimedBodyPart
			}
		}
	}

	// Build candidate list (skip amputated parts).
	type candidate struct {
		name   string
		weight int
	}
	var candidates []candidate
	totalWeight := 0

	if entity.HasComponent(component.HitLocation) {
		hlc := entity.GetComponent(component.HitLocation).(*component.HitLocationComponent)
		for name, part := range bc.Parts {
			if part.Amputated {
				continue
			}
			w := hlc.Weights[name]
			if w <= 0 {
				w = 1
			}
			candidates = append(candidates, candidate{name, w})
			totalWeight += w
		}
	} else {
		for name, part := range bc.Parts {
			if part.Amputated {
				continue
			}
			candidates = append(candidates, candidate{name, 1})
			totalWeight++
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	r := rand.Intn(totalWeight)
	for _, c := range candidates {
		r -= c.weight
		if r < 0 {
			return c.name
		}
	}
	return candidates[len(candidates)-1].name
}

// armorSP returns the StoppingPower of armor equipped on the body part that was hit.
func armorSP(entity *ecs.Entity, partName string) int {
	sp := 0

	// Check entity-level natural SP (e.g. thick_skin skill grants NaturalSP).
	if entity.HasComponent(component.Stats) {
		sc := entity.GetComponent(component.Stats).(*component.StatsComponent)
		sp += sc.NaturalSP
	}

	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
		// Equipped is keyed by body part name.
		item := inv.Equipped[partName]
		if item != nil && item.HasComponent(component.Armor) {
			ac := item.GetComponent(component.Armor).(*rlcomponents.ArmorComponent)
			sp += ac.StoppingPower
		}
	}

	return sp
}

// applyPartDamage applies damage to a body part, handling damage type resistances/weaknesses,
// and setting broken/amputated flags. Returns broken, amputated, killed.
func applyPartDamage(
	bc *rlcomponents.BodyComponent,
	partName string,
	damage int,
	damageType string,
	entity *ecs.Entity,
) (broken, amputated, killed bool) {
	// Apply resistance/weakness from entity stats.
	if entity.HasComponent(component.Stats) {
		sc := entity.GetComponent(component.Stats).(*component.StatsComponent)
		if slices.Contains(sc.Resistances, damageType) {
			damage /= 2
		} else if slices.Contains(sc.Weaknesses, damageType) {
			damage *= 2
		}
	}

	part, ok := bc.Parts[partName]
	if !ok {
		return
	}
	part.HP -= damage
	overkill := part.HP < 0
	if part.HP <= 0 {
		part.HP = 0
		if !part.Broken {
			part.Broken = true
			broken = true
			if part.KillsWhenBroken {
				killed = true
			}
		}
		if overkill && !part.Amputated && part.KillsWhenAmputated {
			part.Amputated = true
			amputated = true
			killed = true
		}
	}
	bc.Parts[partName] = part
	return
}
