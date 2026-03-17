package system

import (
	"math/rand"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/path"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/world"
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

			if rlentity.HandleDeath(entity) {
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
				rlentity.Face(entity, deltaX, deltaY)
			}

			//Hostile AI
			if entity.HasComponent("HostileAI") {
				hc := entity.GetComponent("HostileAI").(*component.HostileAIComponent)
				deltaX := 0
				deltaY := 0
				z := pc.GetZ()

				//Scan around for food to the best my vision allows me.
				var nearby []*ecs.Entity
				level.GetEntitiesAround(pc.GetX(), pc.GetY(), z, hc.SightRange, hc.SightRange, &nearby)
				if len(nearby) > 0 {
					closest := entity
					distance := 999999.0
					for e := range nearby {
						if nearby[e] != entity {
							if !rlcombat.IsFriendly(entity, nearby[e]) {
								foodPC := nearby[e].GetComponent("Position").(*component.PositionComponent)
								if nearby[e].HasComponent("Food") && !nearby[e].HasComponent("Dead") {
									from := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
									to := level.Level.GetTilePtr(foodPC.GetX(), foodPC.GetY(), z)
									if from != nil && to != nil {
										tDistance := from.PathEstimatedCost(to)
										if tDistance < distance {
											closest = nearby[e]
											distance = tDistance
										}
									}
								}
							}
						}
					}

					if closest != entity {
						foodPC := closest.GetComponent("Position").(*component.PositionComponent)
						hc.TargetX = foodPC.GetX()
						hc.TargetY = foodPC.GetY()
						from := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
						to := level.Level.GetTilePtr(hc.TargetX, hc.TargetY, z)
						if from != nil && to != nil {
							steps, _, _ := path.Path(from, to)
							if len(steps) > 0 {
								t := steps[0].(*world.Tile)
								tx, ty, _ := t.Coords()
								if pc.GetX() < tx {
									deltaX = 1
								}
								if pc.GetX() > tx {
									deltaX = -1
								}
								if pc.GetY() < ty {
									deltaY = 1
								}
								if pc.GetY() > ty {
									deltaY = -1
								}
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
					entityHit := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY, z)
					if entityHit != nil && entityHit != entity {
						hit(level, entity, entityHit)
						rlentity.Eat(entity, entityHit)
					}
				}
				rlentity.Face(entity, deltaX, deltaY)
			}

			//Defensive AI
			if entity.HasComponent("DefensiveAI") {
				aic := entity.GetComponent("DefensiveAI").(*component.DefensiveAIComponent)

				if aic.Attacked {
					z := pc.GetZ()
					entityHit := level.GetSolidEntityAt(aic.AttackerX, aic.AttackerY, z)

					if entityHit == nil {
						aic.Attacked = false
					} else {
						hit(level, entity, entityHit)
					}

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

					rlentity.Face(entity, deltaX, deltaY)
				}
			}
		}
	}

	return nil
}
