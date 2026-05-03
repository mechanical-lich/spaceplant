package entityhelpers

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/ecs"
	spcombat "github.com/mechanical-lich/spaceplant/internal/combat"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

func HealBodyParts(entity *ecs.Entity, amount int) {
	if !entity.HasComponent(component.Body) {
		return
	}
	bc := entity.GetComponent(component.Body).(*component.BodyComponent)
	var damaged []string
	for name, part := range bc.Parts {
		if !part.Amputated && part.HP < part.MaxHP {
			damaged = append(damaged, name)
		}
	}
	if len(damaged) == 0 {
		return
	}
	perPart := amount / len(damaged)
	remainder := amount % len(damaged)
	for i, name := range damaged {
		part := bc.Parts[name]
		heal := perPart
		if i < remainder {
			heal++
		}
		part.HP += heal
		if part.HP > part.MaxHP {
			part.HP = part.MaxHP
		}
		if part.HP > 0 && part.Broken {
			part.Broken = false
		}
		bc.Parts[name] = part
	}
}

// Hit performs a melee attack using the AAG combat system, then adds the
// AttackComponent visual on a hit.
// Returns true if the attack landed.
func Hit(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity) bool {
	landed := spcombat.Hit(l, entity, entityHit)
	if landed {
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			hitX, hitY := HitTile(entity, entityHit)
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
		}
	}
	return landed
}

// HitRanged performs a ranged attack with a specific weapon and CS bonus/penalty.
// Pass the exact weapon being fired so CS modifier, Pen, and damage type are
// always sourced from the correct item regardless of other equipped weapons.
func HitRanged(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity, weapon *component.WeaponComponent, csBonus int) bool {
	landed := spcombat.HitRanged(l, entity, entityHit, weapon, csBonus)
	if landed {
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			hitX, hitY := HitTile(entity, entityHit)
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
		}
	}
	return landed
}

// HitRangedTargeted performs a ranged attack biased toward a specific body part.
func HitRangedTargeted(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity, weapon *component.WeaponComponent, csBonus int, aimedBodyPart string) bool {
	landed := spcombat.HitRangedTargeted(l, entity, entityHit, weapon, csBonus, aimedBodyPart)
	if landed {
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			hitX, hitY := HitTile(entity, entityHit)
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
		}
	}
	return landed
}

// HitWithPen performs an attack with a Penetration override (e.g. for unarmed
// special attacks). penOverride < 0 means use the weapon/bare-hands default.
func HitWithPen(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity, pen int) bool {
	landed := spcombat.HitWithPen(l, entity, entityHit, pen)
	if landed {
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			hitX, hitY := HitTile(entity, entityHit)
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
		}
	}
	return landed
}

// ResistCheck runs a stat-based resistance check on entityHit.
func ResistCheck(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity, dc int, stat string) bool {
	_ = l
	passed := spcombat.ResistCheck(entityHit, dc, stat)
	if !passed && entityHit != entity && !entityHit.HasComponent(component.Dead) {
		hitX, hitY := HitTile(entity, entityHit)
		entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
	}
	return passed
}

// HitTile returns the world tile on the target's footprint closest to the attacker.
func HitTile(attacker, target *ecs.Entity) (int, int) {
	return rlentity.HitTile(attacker, target)
}

// LegWounded returns true if the entity has any leg-role body part below 50% HP.
// Used to double movement action cost.
func LegWounded(entity *ecs.Entity) bool {
	return rlentity.LegWounded(entity)
}

// LegPenaltyCost returns the extra action cost to add due to broken or amputated leg parts.
// Each broken leg adds half the base move cost; each amputated leg adds the full base move cost.
func LegPenaltyCost(entity *ecs.Entity, baseCost int) int {
	return rlentity.LegPenaltyCost(entity, baseCost)
}

// Move attempts to move an entity using rlentity.Move, with MassiveComponent handling.
func Move(entity *ecs.Entity, level *world.Level, deltaX int, deltaY int) bool {
	pc := entity.GetComponent("Position").(*component.PositionComponent)
	z := pc.GetZ()
	blocker := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY, z)

	if blocker != nil && blocker != entity {
		// Massive entities destroy non-massive blockers
		if entity.HasComponent("MassiveComponent") && !blocker.HasComponent("MassiveComponent") {
			blocker.AddComponent(&component.DeadComponent{})
		}
	}

	return rlentity.Move(entity, level.Level, deltaX, deltaY, 0)
}
