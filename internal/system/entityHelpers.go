package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// hit performs a melee attack using rlcombat.Hit, then adds the AttackComponent visual.
// The FX is placed on the target's footprint tile closest to the attacker.
func hit(l *world.Level, entity *ecs.Entity, entityHit *ecs.Entity) {
	if rlcombat.Hit(l.Level, entity, entityHit, true) {
		if entityHit != entity && !entityHit.HasComponent(component.Dead) {
			hitX, hitY := hitTile(entity, entityHit)
			entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0, X: hitX, Y: hitY})
		}
	}
}

// hitTile returns the world tile on the target's footprint closest to the attacker.
func hitTile(attacker, target *ecs.Entity) (int, int) {
	apc := attacker.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	tpc := target.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	tw, th := 1, 1
	if target.HasComponent(rlcomponents.Size) {
		sc := target.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		if sc.Width > 0 {
			tw = sc.Width
		}
		if sc.Height > 0 {
			th = sc.Height
		}
	}
	startX := tpc.GetX() - tw/2
	startY := tpc.GetY() - th/2
	return max(startX, min(apc.GetX(), startX+tw-1)),
		max(startY, min(apc.GetY(), startY+th-1))
}

// move attempts to move an entity using rlentity.Move, with MassiveComponent handling.
func move(entity *ecs.Entity, level *world.Level, deltaX int, deltaY int) bool {
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
