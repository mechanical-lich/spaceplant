package system

import (
	"errors"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type LightSystem struct {
}

func (s LightSystem) UpdateSystem(data any) error {
	return nil
}

func (s LightSystem) Requires() []ecs.ComponentType {
	return nil
}

func (s LightSystem) ClearLights(l *world.Level, z int) {
	for x := 0; x < l.Width; x++ {
		for y := 0; y < l.Height; y++ {
			t := l.Level.GetTilePtr(x, y, z)
			if t != nil {
				t.LightLevel = 255
			}
		}
	}
}

// LightSystem .
func (s LightSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	if entity.HasComponent("LightComponent") && entity.HasComponent("Position") {
		l, ok := levelInterface.(*world.Level)
		if !ok {
			return errors.New("invalid level type")
		}

		lc := entity.GetComponent("LightComponent").(*component.LightComponent)

		pc := entity.GetComponent("Position").(*component.PositionComponent)
		z := pc.GetZ()
		for x := pc.GetX() - lc.Radius; x < pc.GetX()+lc.Radius; x++ {
			for y := pc.GetY() - lc.Radius; y < pc.GetY()+lc.Radius; y++ {
				if rlfov.Los(l.Level, x, y, pc.GetX(), pc.GetY(), z) {
					dist := utility.Distance(x, y, pc.GetX(), pc.GetY())
					if dist == 0 {
						dist = 1
					}
					// Avoid divide-by-zero if Brightness wasn't set in blueprints
					brightness := lc.Brightness
					if brightness == 0 {
						brightness = 1
					}
					t := l.Level.GetTilePtr(x, y, z)
					if t != nil {
						t.LightLevel -= 255 - 255*dist/brightness
						if t.LightLevel > 255 {
							t.LightLevel = 255
						}
						if t.LightLevel <= 0 {
							t.LightLevel = 0
						}
					}
				}
			}
		}
	}
	return nil
}
