package system

import (
	"fmt"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/dice"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/message"
	log "github.com/sirupsen/logrus"
)

func hit(l *level.Level, entity *ecs.Entity, entityHit *ecs.Entity) {
	if entityHit != entity {
		//Faction check
		if entityHit.HasComponent("Description") && entity.HasComponent("Description") {
			hitDc := entityHit.GetComponent("Description").(*component.DescriptionComponent)
			dc := entity.GetComponent("Description").(*component.DescriptionComponent)

			hitPc := entityHit.GetComponent("Position").(*component.PositionComponent)
			pc := entity.GetComponent("Position").(*component.PositionComponent)

			oldX := pc.GetX()
			oldY := pc.GetY()
			if dc.Faction != "none" && dc.Faction != "" {
				//Swap position
				if dc.Faction == hitDc.Faction {
					l.PlaceEntity(hitPc.GetX(), hitPc.GetY(), entity)
					l.PlaceEntity(oldX, oldY, entityHit)
					return
				}
			}
		}

		// Attack it
		if entityHit.HasComponent("Health") && entityHit.HasComponent("Stats") && entity.HasComponent("Stats") {
			sc := entity.GetComponent("Stats").(*component.StatsComponent)
			hitSc := entityHit.GetComponent("Stats").(*component.StatsComponent)

			// Apply weapons
			attackMod := 0
			attackDice := sc.BasicAttackDice

			if entity.HasComponent("Inventory") {
				inventory := entity.GetComponent("Inventory").(*component.InventoryComponent)
				attackMod += inventory.GetAttackModifier()
				d := inventory.GetAttackDice()
				if d != "" {
					attackDice = d
				}
			}

			//Roll to hit
			d := fmt.Sprintf("1d20+%d", getModifier(sc.Dex))
			roll, err := dice.ParseDiceRequest(d)
			if err == nil {
				if roll.Result > hitSc.AC {
					damage := 0
					d = fmt.Sprintf("%s+%d", attackDice, getModifier(sc.Str)+attackMod)
					roll, err = dice.ParseDiceRequest(d)
					if err == nil {
						damage = roll.Result
					} else {
						log.Error("Error rolling dice: ", d)
					}

					//Create visual of attack
					entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0})

					// Apply defenses
					if entityHit.HasComponent("Inventory") {
						inventory := entityHit.GetComponent("Inventory").(*component.InventoryComponent)
						damage -= inventory.GetDefenseModifier()
						if damage < 0 {
							damage = 0
						}
					}

					// Apply damage
					ehc := entityHit.GetComponent("Health").(*component.HealthComponent)
					if ehc.Health > 0 {
						ehc.Health -= damage
						if entityHit.HasComponent("Description") && entity.HasComponent("Description") {
							hitDc := entityHit.GetComponent("Description").(*component.DescriptionComponent)
							dc := entity.GetComponent("Description").(*component.DescriptionComponent)
							message.AddMessage(fmt.Sprint(dc.Name+" hit "+hitDc.Name+" for ", damage))
						}
					}
				} else {
					// Missed
					if entityHit.HasComponent("Description") && entity.HasComponent("Description") {
						hitDc := entityHit.GetComponent("Description").(*component.DescriptionComponent)
						dc := entity.GetComponent("Description").(*component.DescriptionComponent)
						message.AddMessage(fmt.Sprint(dc.Name + " missed " + hitDc.Name))
					}
				}
			} else {
				log.Error("Error rolling dice: ", d)

			}
		}

		// Trigger their defenses
		if entityHit.HasComponent("DefensiveAI") {
			daic := entityHit.GetComponent("DefensiveAI").(*component.DefensiveAIComponent)
			pc := entity.GetComponent("Position").(*component.PositionComponent)
			daic.Attacked = true
			daic.AttackerX = pc.GetX()
			daic.AttackerY = pc.GetY()
		}

		if !entityHit.HasComponent("Alerted") {
			entityHit.AddComponent(&component.AlertedComponent{Duration: 120})
		}

		if entity.HasComponent("Poisonous") {
			if !entityHit.HasComponent("Poisoned") {
				poisonousComponent := entity.GetComponent("Poisonous").(*component.PoisonousComponent)
				entityHit.AddComponent(&component.PoisonedComponent{Duration: poisonousComponent.Duration})
			}
		}
	}

}

// Returns true if successfully ate.
func eat(entity *ecs.Entity, entityHit *ecs.Entity) bool {
	if entityHit != entity {
		if entityHit.HasComponent("Food") {
			fc := entityHit.GetComponent("Food").(*component.FoodComponent)
			fc.Amount--
			return true
		}
	}
	return false
}

func face(entity *ecs.Entity, deltaX int, deltaY int) {
	dc := entity.GetComponent("Direction").(*component.DirectionComponent)
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

func handleDeath(entity *ecs.Entity) bool {
	if entity.HasComponent("Health") {
		hc := entity.GetComponent("Health").(*component.HealthComponent)
		if hc.Health <= 0 {
			entity.AddComponent(&component.DeadComponent{})

			return true
		}
	}
	return false
}

// Returns true if a solid entity is in the way.
func move(entity *ecs.Entity, level *level.Level, deltaX int, deltaY int) bool {
	pc := entity.GetComponent("Position").(*component.PositionComponent)
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
			entityHit.AddComponent(&component.DeadComponent{})
		}
	}

	return true
}

func getModifier(stat int) int {
	return (stat - 10) / 2
}
