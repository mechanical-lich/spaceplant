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
	return hitCore(level, attacker, defender, nil, -1, 0)
}

// HitWithPen resolves an attack with an optional Pen override.
// If penOverride >= 0 it replaces the weapon's Penetration value (used by
// special unarmed attacks like hands_only).
func HitWithPen(level *world.Level, attacker, defender *ecs.Entity, penOverride int) bool {
	return hitCore(level, attacker, defender, nil, penOverride, 0)
}

// HitRanged resolves a ranged attack with a specific weapon and CS bonus/penalty.
// weaponOverride pins which weapon is used for CS modifier, Pen, and damage type,
// preventing non-deterministic map iteration from selecting the wrong equipped item.
// csBonus is added on top of the weapon's CombatSkillModifier.
func HitRanged(level *world.Level, attacker, defender *ecs.Entity, weaponOverride *rlcomponents.WeaponComponent, csBonus int) bool {
	return hitCore(level, attacker, defender, weaponOverride, -1, csBonus)
}

// hitCore is the shared hit-resolution engine.
// weaponOverride, when non-nil, is used in place of equippedWeapon(attacker).
// penOverride < 0 means use the weapon / bare-hands Pen. csBonus is added to CS
// after the weapon's CombatSkillModifier (range bands, aimed shot, etc.).
func hitCore(level *world.Level, attacker, defender *ecs.Entity, weaponOverride *rlcomponents.WeaponComponent, penOverride, csBonus int) bool {
	apc := attacker.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	dpc := defender.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

	attackerName := rlentity.GetName(attacker)
	defenderName := rlentity.GetName(defender)

	// --- Determine effective CombatSkill ---
	cs := 30 // fallback if no stats
	if attacker.HasComponent(component.Stats) {
		sc := attacker.GetComponent(component.Stats).(*component.StatsComponent)
		cs = sc.CS
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
		weaponName = damageType // good enough for messages
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
	partName := pickBodyPart(defender)

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
		Broken:       broken,
		Amputated:    amputated,
	})

	return true
}

// CoolCheck resolves a Cool-based resistance roll against a difficulty value.
// Returns true if the entity resists (succeeds). dc is addded difficulty (higher = harder).
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

// --- helpers ---

// equippedWeapon returns the first equipped WeaponComponent found on the entity, or nil.
func equippedWeapon(entity *ecs.Entity) *rlcomponents.WeaponComponent {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				return item.GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
			}
		}
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
		if inv.RightHand != nil && inv.RightHand.HasComponent(component.Weapon) {
			return inv.RightHand.GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
		}
		if inv.LeftHand != nil && inv.LeftHand.HasComponent(component.Weapon) {
			return inv.LeftHand.GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
		}
	}
	return nil
}

// pickBodyPart selects a body part name using HitLocationComponent weights.
// Falls back to equal-weight selection if HitLocationComponent is absent.
func pickBodyPart(entity *ecs.Entity) string {
	if !entity.HasComponent(rlcomponents.Body) {
		return ""
	}
	bc := entity.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)

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
	if part.HP <= 0 {
		part.HP = 0
		if !part.Broken {
			part.Broken = true
			broken = true
			if part.KillsWhenBroken {
				killed = true
			}
		}
		if !part.Amputated && part.KillsWhenAmputated {
			part.Amputated = true
			amputated = true
			killed = true
		}
	}
	bc.Parts[partName] = part
	return
}
