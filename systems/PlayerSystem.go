package systems

import (
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/components"
	"github.com/mechanical-lich/spaceplant/level"
)

type PlayerSystem struct {
}

// PlayerSystem .
func (s *PlayerSystem) Update(levelInterface interface{}, entity *ecs.Entity) error {
	level := levelInterface.(*level.Level)

	if entity.HasComponent("PlayerComponent") {
		if entity.HasComponent("MyTurnComponent") {
			pc := entity.GetComponent("PositionComponent").(*components.PositionComponent)
			dc := entity.GetComponent("DirectionComponent").(*components.DirectionComponent)
			playerComponent := entity.GetComponent("PlayerComponent").(*components.PlayerComponent)
			command := playerComponent.PopCommand()

			// pX := pc.GetX()
			// pY := pc.GetY()

			deltaX := 0
			deltaY := 0
			switch command {
			case "W":
				deltaY--

				dc.Direction = 2
			case "S":
				deltaY++

				dc.Direction = 1
			case "A":
				deltaX--

				dc.Direction = 3
			case "D":
				deltaX++

				dc.Direction = 0
			case "F":
				direction := playerComponent.PopCommand()
				if direction == "" {
					playerComponent.AddMessage("Wasn't given a direction to shoot!")
				} else {
					playerComponent.AddMessage("Shoot in the " + direction + " direction!")
				}
			}
			if move(entity, level, deltaX, deltaY) {
				entityHit := level.GetSolidEntityAt(pc.GetX()+deltaX, pc.GetY()+deltaY)
				if entityHit != nil && entityHit != entity {
					if entityHit != entity {
						hit(level, entity, entityHit)
						//eat(entity, entityHit)
					}
				}
			}
		}
	}

	return nil
}
