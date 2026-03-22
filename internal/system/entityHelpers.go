package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// hit performs a melee attack using rlcombat.Hit, then adds the AttackComponent visual.
func hit(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity) {
	// Add visual FX before the hit (so it shows even on miss via faction swap)
	// if entityHit != entity && !rlcombat.IsFriendly(entity, entityHit) {
	// 	entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0})
	// }
	if rlcombat.Hit(l.Level, entity, entityHit, true) {
		// Add visual FX after the hit (so it shows on hit even if friendly swap)
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0})
		}
	}
}

// move attempts to move an entity using rlentity.Move, with MassiveComponent handling.
func move(entity *ecs.Entity, level *world.Level, deltaX int, deltaY int) bool {
	pc := entity.GetComponent("Position").(*component.PositionComponent)
	z := pc.GetZ()
	blocker := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY, z)

	if blocker != nil {
		// Massive entities destroy non-massive blockers
		if entity.HasComponent("MassiveComponent") && !blocker.HasComponent("MassiveComponent") {
			blocker.AddComponent(&component.DeadComponent{})
		}
	}

	return rlentity.Move(entity, level.Level, deltaX, deltaY, 0)
}
