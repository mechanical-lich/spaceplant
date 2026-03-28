package system

import (
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/path"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

func getRandom(low int, high int) int {
	return (rand.Intn((high - low))) + low
}

type AISystem struct {
	Watcher *ecs.Entity
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
			isAI := entity.HasComponent("WanderAI") || entity.HasComponent("HostileAI") || entity.HasComponent("DefensiveAI")
			if !isAI {
				return nil
			}
			entity.AddComponent(rlcomponents.GetTurnTaken())
			pc := entity.GetComponent("Position").(*component.PositionComponent)
			actionCost := energy.CostMove

			if rlentity.HandleDeath(entity) {
				rlentity.CheckDeathAnnouncement(s.Watcher, entity, level.Level)
				return nil
			}

			//Wander AI
			if entity.HasComponent("WanderAI") {
				deltaX := getRandom(-1, 2)
				deltaY := 0
				if deltaX == 0 {
					deltaY = getRandom(-1, 2)
				}
				if entityhelpers.Move(entity, level, deltaX, deltaY) {
					destTile := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), pc.GetZ())
					if destTile != nil {
						actionCost = rlenergy.MoveCost(destTile, energy.CostMove)
					}
				}
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
										tDistance := level.Level.PathEstimate(from.Idx, to.Idx)
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

						// Start pathfinding from the footprint tile closest to the target.
						fromX, fromY := pc.GetX(), pc.GetY()
						var graph path.Graph = level.Level
						if entity.HasComponent(rlcomponents.Size) {
							sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
							w, h := sc.Width, sc.Height
							if w > 1 || h > 1 {
								graph = &rlworld.SizedGraph{Level: level.Level, Width: w, Height: h, Entity: entity}
								startX := fromX - w/2
								startY := fromY - h/2
								fromX = max(startX, min(hc.TargetX, startX+w-1))
								fromY = max(startY, min(hc.TargetY, startY+h-1))
							}
						}

						from := level.Level.GetTilePtr(fromX, fromY, z)
						to := level.Level.GetTilePtr(hc.TargetX, hc.TargetY, z)
						if from != nil && to != nil {
							steps, _, _ := path.Path(graph, from.Idx, to.Idx)
							hc.Path = append(hc.Path[:0], steps...)
							if len(steps) > 1 {
								// Derive direction from the path itself (always cardinal).
								s0 := level.Level.GetTilePtrIndex(steps[0])
								s1 := level.Level.GetTilePtrIndex(steps[1])
								sx, sy, _ := s0.Coords()
								nx, ny, _ := s1.Coords()
								deltaX = nx - sx
								deltaY = ny - sy
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

				if entityhelpers.Move(entity, level, deltaX, deltaY) {
					var blockers []*ecs.Entity
					rlentity.FootprintBlockers(entity, level, pc.GetX()+deltaX, pc.GetY()+deltaY, z, &blockers)
					if len(blockers) > 0 {
						actionCost = energy.CostAttack
						for _, entityHit := range blockers {
							entityhelpers.Hit(level, entity, entityHit)
							rlentity.Eat(entity, entityHit)
						}
					} else {
						destTile := level.Level.GetTilePtr(pc.GetX(), pc.GetY(), z)
						if destTile != nil {
							actionCost = rlenergy.MoveCost(destTile, energy.CostMove)
						}
					}
				}
				rlentity.Face(entity, deltaX, deltaY)
			}

			//Defensive AI
			if entity.HasComponent("DefensiveAI") {
				aic := entity.GetComponent("DefensiveAI").(*component.DefensiveAIComponent)

				if aic.Attacked {
					actionCost = energy.CostAttack
					z := pc.GetZ()
					entityHit := level.GetSolidEntityAt(aic.AttackerX, aic.AttackerY, z)

					if entityHit == nil {
						aic.Attacked = false
					} else {
						entityhelpers.Hit(level, entity, entityHit)
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

			rlenergy.SetActionCost(entity, actionCost)
		}
	}

	return nil
}
