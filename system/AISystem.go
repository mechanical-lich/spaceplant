package system

import (
	"math/rand"

	"github.com/beefsack/go-astar"
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
	world "github.com/mechanical-lich/spaceplant/level"
)

func getRandom(low int, high int) int {
	return (rand.Intn((high - low))) + low
}

type AISystem struct {
}

// PlayerSystem .
func (s AISystem) Update(levelInterface interface{}, entity *ecs.Entity) error {
	level := levelInterface.(*world.Level)
	if !entity.HasComponent("DeadComponent") {
		if entity.HasComponent("MyTurnComponent") {

			pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)

			if handleDeath(entity) {
				return nil
			}

			//Wander AI
			if entity.HasComponent("WanderAIComponent") {
				deltaX := getRandom(-1, 2)
				deltaY := 0
				if deltaX == 0 {
					deltaY = getRandom(-1, 2)
				}
				move(entity, level, deltaX, deltaY)
				face(entity, deltaX, deltaY)
			}

			//Hostile AI
			if entity.HasComponent("HostileAIComponent") {
				hc := entity.GetComponent("HostileAIComponent").(*component.HostileAIComponent)
				deltaX := 0
				deltaY := 0

				//Scan around for food to the best my vision allows me.
				nearby := level.GetEntitiesAround(pc.GetX(), pc.GetY(), hc.SightRange, hc.SightRange)
				if len(nearby) > 0 {
					closest := entity
					distance := 999999.0
					for e := range nearby {
						if nearby[e] != entity {
							friendly := false
							if entity.HasComponent("DescriptionComponent") {
								if nearby[e].HasComponent("DescriptionComponent") {
									myDC := entity.GetComponent("DescriptionComponent").(*component.DescriptionComponent)
									hitDC := nearby[e].GetComponent("DescriptionComponent").(*component.DescriptionComponent)

									if myDC.Faction != "none" && myDC.Faction != "" {
										if myDC.Faction == hitDC.Faction {
											friendly = true

										}
									}
								}
							}
							if !friendly {
								foodPC := nearby[e].GetComponent("PositionComponent").(*component.PositionComponent)
								if nearby[e].HasComponent("FoodComponent") && !nearby[e].HasComponent("DeadComponent") {
									tDistance := level.GetTileAt(pc.GetX(), pc.GetY()).PathEstimatedCost(level.GetTileAt(foodPC.GetX(), foodPC.GetY()))
									if tDistance < distance {

										closest = nearby[e]
										distance = tDistance
									}
								}
							}
						}
					}

					if closest != entity {
						foodPC := closest.GetComponent("PositionComponent").(*component.PositionComponent)
						hc.TargetX = foodPC.GetX()
						hc.TargetY = foodPC.GetY()
						steps, _, _ := astar.Path(level.GetTileAt(pc.GetX(), pc.GetY()), level.GetTileAt(hc.TargetX, hc.TargetY))
						if len(steps) > 0 {
							t := steps[0].(*world.Tile)
							if pc.GetX() < t.X {
								deltaX = 1
							}

							if pc.GetX() > t.X {
								deltaX = -1
							}

							if pc.GetY() < t.Y {
								deltaY = 1
							}

							if pc.GetY() > t.Y {
								deltaY = -1
							}
						}
					}
				}

				//Found nothing, wander
				if deltaX == 0 && deltaY == 0 {
					deltaX = getRandom(-1, 2)
					deltaY = 0
					if deltaX == 0 {
						deltaY = getRandom(-1, 2)
					}
				}

				if move(entity, level, deltaX, deltaY) {
					entityHit := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY)
					if entityHit != nil && entityHit != entity {
						if entityHit != entity {
							hit(level, entity, entityHit)
							eat(entity, entityHit)
						}
					}
				}
				face(entity, deltaX, deltaY)
			}

			//Defensive AI
			if entity.HasComponent("DefensiveAIComponent") {
				aic := entity.GetComponent("DefensiveAIComponent").(*component.DefensiveAIComponent)

				if aic.Attacked {
					entityHit := level.GetSolidEntityAt(aic.AttackerX, aic.AttackerY)

					if entityHit == nil {
						// No attacker there.
						aic.Attacked = false
					} else {
						// Hit the attacker back.
						hit(level, entity, entityHit)
					}

					// Point where you attack
					deltaX := 0
					deltaY := 0
					if pc.GetX() < aic.AttackerX {
						deltaX = 1
					}

					if pc.GetX() > aic.AttackerX {
						deltaX = -1
					}

					if pc.GetY() < aic.AttackerY {
						deltaY = 1
					}

					if pc.GetY() > aic.AttackerY {
						deltaY = -1
					}

					face(entity, deltaX, deltaY)
				}
			}
		}
	}

	return nil
}
