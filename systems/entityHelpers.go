package systems

import (
	"github.com/mechanical-lich/game-engine/entity"
	"github.com/mechanical-lich/spaceplant/components"
	"github.com/mechanical-lich/spaceplant/level"
)

func hit(entity *entity.Entity, entityHit *entity.Entity) {
	if entityHit != entity {
		//Attack it
		if entityHit.HasComponent("HealthComponent") {
			damage := 1
			if entityHit.HasComponent("DamageComponent") {
				dc := entityHit.GetComponent("DamageComponent").(*components.DamageComponent)
				damage = dc.Amount
			}
			ehc := entityHit.GetComponent("HealthComponent").(*components.HealthComponent)
			if ehc.Health > 0 {
				ehc.Health -= damage
			}
		}

		//Create visual of attack
		entityHit.AddComponent(&components.AttackComponent{SpriteX: 0, SpriteY: 0})

		// Trigger their defenses
		if entityHit.HasComponent("DefensiveAIComponent") {
			daic := entityHit.GetComponent("DefensiveAIComponent").(*components.DefensiveAIComponent)
			pc := entity.GetComponent("PositionComponent").(*components.PositionComponent)
			daic.Attacked = true
			daic.AttackerX = pc.GetX()
			daic.AttackerY = pc.GetY()
		}

		if !entityHit.HasComponent("AlertedComponent") {
			entityHit.AddComponent(&components.AlertedComponent{Duration: 120})
		}

		if entity.HasComponent("PoisonousComponent") {
			if !entityHit.HasComponent("PoisonedComponent") {
				poisonousComponent := entity.GetComponent("PoisonousComponent").(*components.PoisonousComponent)
				entityHit.AddComponent(&components.PoisonedComponent{Duration: poisonousComponent.Duration})
			}
		}
	}
}

// Returns true if successfully ate.
func eat(entity *entity.Entity, entityHit *entity.Entity) bool {
	if entityHit != entity {
		if entityHit.HasComponent("FoodComponent") {
			fc := entityHit.GetComponent("FoodComponent").(*components.FoodComponent)
			fc.Amount--
			return true
		}
	}
	return false
}

func face(entity *entity.Entity, deltaX int, deltaY int) {
	dc := entity.GetComponent("DirectionComponent").(*components.DirectionComponent)
	if deltaY > 0 {
		dc.Direction = 1
	}
	if deltaY < 0 {
		dc.Direction = 2
	}
	if deltaX < 0 {
		dc.Direction = 3
	}
	if deltaX > 0 {
		dc.Direction = 0
	}
}

func handleDeath(entity *entity.Entity) bool {
	if entity.HasComponent("HealthComponent") {
		hc := entity.GetComponent("HealthComponent").(*components.HealthComponent)
		if hc.Health <= 0 {
			entity.AddComponent(&components.DeadComponent{})

			return true
		}
	}
	return false
}

// Returns true if a solid entity is in the way.
func move(entity *entity.Entity, level *level.Level, deltaX int, deltaY int) bool {
	pc := entity.GetComponent("PositionComponent").(*components.PositionComponent)
	entityHit := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY)
	if entityHit == nil {
		tile := level.GetTileAt(pc.GetX()+deltaX, pc.GetY()+deltaY)
		if tile == nil {
		} else if !tile.Solid {
			level.PlaceEntity(pc.GetX()+deltaX, pc.GetY()+deltaY, entity)
		}
		return false
	} else {
		//If we are massive we destroy what we moved into if it's not also massive.
		if entity.HasComponent("MassiveComponent") && !entityHit.HasComponent("MassiveComponent") {
			entityHit.AddComponent(&components.DeadComponent{})
		}
	}

	return true
}
