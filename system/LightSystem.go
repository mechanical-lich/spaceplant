package system

import (
	"errors"

	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/utility"
)

type LightSystem struct {
}

func (s LightSystem) ClearLights(l *level.Level) {
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			l.GetTileAt(x, y).Light = 255
		}
	}
}

// LightSystem .
func (s LightSystem) Update(levelInterface interface{}, entity *ecs.Entity) error {
	if entity.HasComponent("LightComponent") && entity.HasComponent("PositionComponent") {
		l, ok := levelInterface.(*level.Level)
		if !ok {
			return errors.New("invalid level type")
		}

		lc := entity.GetComponent("LightComponent").(*component.LightComponent)

		pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
		for x := pc.GetX() - lc.Radius; x < pc.GetX()+lc.Radius; x++ {
			for y := pc.GetY() - lc.Radius; y < pc.GetY()+lc.Radius; y++ {
				if level.Los(x, y, pc.GetX(), pc.GetY(), l) {
					dist := utility.Distance(x, y, pc.GetX(), pc.GetY())
					if dist == 0 {
						dist = 1
					}
					t := l.GetTileAt(x, y)
					if t != nil {
						t.Light -= 255 - 255*dist/lc.Brightness
						if t.Light > 255 {
							t.Light = 255
						}
						if t.Light <= 0 {
							t.Light = 0
						}
					}
				}
			}
		}
	}
	return nil
}
