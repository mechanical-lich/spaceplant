package system

import (
	"math/rand"

	"github.com/beefsack/go-astar"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/component"
	world "github.com/mechanical-lich/spaceplant/level"
)

func getRandom(low int, high int) int {
	return (rand.Intn((high - low))) + low
}

type AISystem struct {
}

func (s *AISystem) UpdateSystem(data any) error {
	return nil
}

func (s *AISystem) Requires() []ecs.ComponentType {
	return nil
}

// AISystem .
func (s *AISystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	level := levelInterface.(*world.Level)
	if !entity.HasComponent("Dead") {
		if entity.HasComponent("MyTurn") {

			pc := entity.GetComponent("Position").(*component.PositionComponent)

			if handleDeath(entity) {
				return nil
			}

			//Wander AI
			if entity.HasComponent("WanderAI") {
				deltaX := getRandom(-1, 2)
				deltaY := 0
				if deltaX == 0 {
					deltaY = getRandom(-1, 2)
				}
				move(entity, level, deltaX, deltaY)
				face(entity, deltaX, deltaY)
			}

			//Hostile AI
			if entity.HasComponent("HostileAI") {
				hc := entity.GetComponent("HostileAI").(*component.HostileAIComponent)
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
							if entity.HasComponent("Description") {
								if nearby[e].HasComponent("Description") {
									myDC := entity.GetComponent("Description").(*component.DescriptionComponent)
									hitDC := nearby[e].GetComponent("Description").(*component.DescriptionComponent)

									if myDC.Faction != "none" && myDC.Faction != "" {
										if myDC.Faction == hitDC.Faction {
											friendly = true

										}
									}
								}
							}
							if !friendly {
								foodPC := nearby[e].GetComponent("Position").(*component.PositionComponent)
								if nearby[e].HasComponent("Food") && !nearby[e].HasComponent("Dead") {
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
						foodPC := closest.GetComponent("Position").(*component.PositionComponent)
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
			if entity.HasComponent("DefensiveAI") {
				aic := entity.GetComponent("DefensiveAI").(*component.DefensiveAIComponent)

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
