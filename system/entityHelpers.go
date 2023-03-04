package system

import (
	"fmt"

	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/dice"
	"github.com/mechanical-lich/spaceplant/level"
)

func hit(l *level.Level, entity *ecs.Entity, entityHit *ecs.Entity) {
	if entityHit != entity {
		//Faction check
		if entityHit.HasComponent("DescriptionComponent") && entity.HasComponent("DescriptionComponent") {
			hitDc := entityHit.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
			dc := entity.GetComponent("DescriptionComponent").(*component.DescriptionComponent)

			hitPc := entityHit.GetComponent("PositionComponent").(*component.PositionComponent)
			pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)

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

		//Attack it
		if entityHit.HasComponent("HealthComponent") && entityHit.HasComponent("StatsComponent") && entity.HasComponent("StatsComponent") {
			sc := entity.GetComponent("StatsComponent").(*component.StatsComponent)
			hitSc := entityHit.GetComponent("StatsComponent").(*component.StatsComponent)

			//Roll to hit
			d := fmt.Sprintf("1d20+%d", getModifier(sc.Dex))
			roll, err := dice.ParseDiceRequest(d)
			if err == nil {
				if roll.Result > hitSc.AC {
					damage := 0
					// TODO Figure out weapon dice.
					d = fmt.Sprintf("%s+%d", sc.BasicAttackDice, getModifier(sc.Str))
					// fmt.Println(d)
					roll, err = dice.ParseDiceRequest(d)
					if err == nil {
						damage = roll.Result
					} else {
						fmt.Println("Error rolling dice: ", d)
					}

					//Create visual of attack
					entityHit.AddComponent(&component.AttackComponent{SpriteX: 0, SpriteY: 0})

					// Apply damage
					ehc := entityHit.GetComponent("HealthComponent").(*component.HealthComponent)
					if ehc.Health > 0 {
						ehc.Health -= damage
						if entityHit.HasComponent("DescriptionComponent") && entity.HasComponent("DescriptionComponent") {
							hitDc := entityHit.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
							dc := entity.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
							fmt.Println(dc.Name+" hit "+hitDc.Name+" for ", damage)
						}
					}
				} else {
					// Missed
					if entityHit.HasComponent("DescriptionComponent") && entity.HasComponent("DescriptionComponent") {
						hitDc := entityHit.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
						dc := entity.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
						fmt.Println(dc.Name + " missed " + hitDc.Name)
					}
				}
			} else {
				fmt.Println("Error rolling dice: ", d)

			}
		}

		// Trigger their defenses
		if entityHit.HasComponent("DefensiveAIComponent") {
			daic := entityHit.GetComponent("DefensiveAIComponent").(*component.DefensiveAIComponent)
			pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
			daic.Attacked = true
			daic.AttackerX = pc.GetX()
			daic.AttackerY = pc.GetY()
		}

		if !entityHit.HasComponent("AlertedComponent") {
			entityHit.AddComponent(&component.AlertedComponent{Duration: 120})
		}

		if entity.HasComponent("PoisonousComponent") {
			if !entityHit.HasComponent("PoisonedComponent") {
				poisonousComponent := entity.GetComponent("PoisonousComponent").(*component.PoisonousComponent)
				entityHit.AddComponent(&component.PoisonedComponent{Duration: poisonousComponent.Duration})
			}
		}
	}
}

// Returns true if successfully ate.
func eat(entity *ecs.Entity, entityHit *ecs.Entity) bool {
	if entityHit != entity {
		if entityHit.HasComponent("FoodComponent") {
			fc := entityHit.GetComponent("FoodComponent").(*component.FoodComponent)
			fc.Amount--
			return true
		}
	}
	return false
}

func face(entity *ecs.Entity, deltaX int, deltaY int) {
	dc := entity.GetComponent("DirectionComponent").(*component.DirectionComponent)
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
	if entity.HasComponent("HealthComponent") {
		hc := entity.GetComponent("HealthComponent").(*component.HealthComponent)
		if hc.Health <= 0 {
			entity.AddComponent(&component.DeadComponent{})

			return true
		}
	}
	return false
}

// Returns true if a solid entity is in the way.
func move(entity *ecs.Entity, level *level.Level, deltaX int, deltaY int) bool {
	pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
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
