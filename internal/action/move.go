package action

import (
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// MoveAction moves the entity one tile in (DeltaX, DeltaY), handling
// bump-attack, door toggle, and interaction on blocked tiles.
type MoveAction struct {
	DeltaX, DeltaY int
}

func (a MoveAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostMove
}

func (a MoveAction) Available(entity *ecs.Entity, level *world.Level) bool {
	if a.DeltaX == 0 && a.DeltaY == 0 {
		return false
	}
	if entityHasSkill(entity, "immobile") {
		return false
	}
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	tx := pc.GetX() + a.DeltaX
	ty := pc.GetY() + a.DeltaY
	tz := pc.GetZ()
	tile := level.Level.GetTilePtr(tx, ty, tz)
	if tile == nil {
		return false
	}
	if !tile.IsSolid() {
		return true
	}
	// Solid tile: still available if there's a live entity to bump (attack/door/interact).
	var candidates []*ecs.Entity
	level.Level.GetEntitiesAt(tx, ty, tz, &candidates)
	for _, e := range candidates {
		if e != entity && !e.HasComponent(rlcomponents.Dead) {
			return true
		}
	}
	return false
}

func (a MoveAction) Execute(entity *ecs.Entity, level *world.Level) error {
	if entityHasSkill(entity, "immobile") {
		rlenergy.SetActionCost(entity, energy.CostMove)
		return nil
	}

	pc := entity.GetComponent("Position").(*component.PositionComponent)
	dc := entity.GetComponent("Direction").(*component.DirectionComponent)
	z := pc.GetZ()

	switch {
	case a.DeltaX > 0:
		dc.Direction = 0
	case a.DeltaX < 0:
		dc.Direction = 3
	case a.DeltaY > 0:
		dc.Direction = 1
	case a.DeltaY < 0:
		dc.Direction = 2
	}

	actionCost := energy.CostMove

	if entityhelpers.Move(entity, level, a.DeltaX, a.DeltaY) {
		// Blocked — prefer non-door entities (monsters) over doors so bumping a
		// monster standing in a doorway attacks rather than toggling the door.
		var candidates []*ecs.Entity
		level.GetEntitiesAt(pc.GetX()+a.DeltaX, pc.GetY()+a.DeltaY, z, &candidates)
		var entityHit, doorHit *ecs.Entity
		for _, e := range candidates {
			if e == entity || e.HasComponent(rlcomponents.Dead) {
				continue
			}
			if e.HasComponent(component.Door) {
				if doorHit == nil {
					doorHit = e
				}
			} else if entityHit == nil {
				entityHit = e
			}
		}
		if entityHit == nil {
			entityHit = doorHit
		}

		if entityHit != nil {
			if rlentity.CheckInteraction(entity, entityHit) {
				// interaction consumed the bump
			} else if entityHit.HasComponent(component.Door) {
				actionCost = energy.CostQuick
				toggleDoor(entity, entityHit)
			} else if rlcombat.IsFriendly(entity, entityHit) {
				actionCost = energy.CostMove
				rlentity.Swap(level.Level, entity, entityHit)
				rlentity.CheckExcuseMe(entity, entityHit)
			} else {
				actionCost = energy.CostAttack
				wc := equippedMeleeWeapon(entity)
				if wc != nil && wc.ActionCost > 0 {
					actionCost = wc.ActionCost
				}
				landed := entityhelpers.Hit(level, entity, entityHit)
				if landed && wc != nil && wc.OnHitCondition != "" && !entityHit.HasComponent(component.Dead) {
					applyOnHitCondition(entityHit, wc)
				}
			}
		}
	} else if a.DeltaX != 0 || a.DeltaY != 0 {
		// Moved successfully — apply terrain-based cost for the destination tile.
		destTile := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
		if destTile != nil {
			actionCost = rlenergy.MoveCost(destTile, energy.CostMove)
			// "Small" skill: ignore elevated terrain costs (e.g. maintenance tunnels).
			if entityHasSkill(entity, "small") && actionCost > energy.CostMove {
				actionCost = energy.CostMove
			}
			// "Trail overgrowth" skill: 10% chance to mark the destination tile overgrown.
			if !destTile.IsAir() && !level.IsOvergrown(pc.GetX(), pc.GetY(), z) && entityHasSkill(entity, "trail_overgrowth") && rand.Intn(100) < 10 {
				level.SetOvergrown(pc.GetX(), pc.GetY(), z, true)
			}
		}
		rlentity.CheckPassOver(entity, level.Level, pc.GetX(), pc.GetY(), z)
		if entityhelpers.LegWounded(entity) {
			actionCost *= 2
		}
		actionCost += entityhelpers.LegPenaltyCost(entity, energy.CostMove)
	}

	rlenergy.SetActionCost(entity, actionCost)
	return nil
}

// toggleDoor opens or closes a door entity, with locked/faction checks.
// It also immediately syncs the appearance sprite so the change is visible
// this tick (door entities are iterated before the player, so DoorSystem
// would otherwise lag one tick behind).
func toggleDoor(actor, doorEntity *ecs.Entity) {
	isPlayer := actor.HasComponent(component.Player)
	door := doorEntity.GetComponent(component.Door).(*component.DoorComponent)
	if door.Open {
		door.Open = false
	} else if door.Locked {
		if isPlayer {
			message.AddMessage("The door is locked.")
		}
		return
	} else if door.OwnedBy != "" {
		if !actor.HasComponent(component.Description) {
			if isPlayer {
				message.AddMessage("Access denied.")
			}
			return
		}
		desc := actor.GetComponent(component.Description).(*component.DescriptionComponent)
		if desc.Faction != door.OwnedBy {
			if isPlayer {
				message.AddMessage("Access denied.")
			}
			return
		}
		door.Open = true
	} else {
		door.Open = true
	}
	rlsystems.SyncDoorAppearance(doorEntity, component.Appearance)
}
