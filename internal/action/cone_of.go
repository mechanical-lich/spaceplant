package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/aoe"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ConeOfFireAction projects a widening cone of fire in the direction the
// entity is currently facing, dealing damage to all entities in range.
//
// Shape at depth 3 (facing right):
//
//	      [1]
//	   [2][2][2]
//	[3][3][3][3][3]
type ConeOfAction struct {
	Params ActionParams
}

func (a ConeOfAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return energy.CostAttack
}

func (a ConeOfAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	return entity.HasComponent(component.Direction)
}

func (a ConeOfAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	ox, oy, z := pc.GetX(), pc.GetY(), pc.GetZ()

	depth := a.Params.Depth
	if depth <= 0 {
		depth = 3
	}
	saveStat := a.Params.SaveStat
	if saveStat == "" {
		saveStat = "dex"
	}
	saveDC := a.Params.SaveDC
	if saveDC <= 0 {
		saveDC = 12
	}
	damageType := a.Params.DamageType
	if damageType == "" {
		damageType = "fire"
	}
	damageDice := a.Params.DamageDice
	if damageDice == "" {
		damageDice = "2d6"
	}

	spread := a.Params.Spread

	fdx, fdy := aoe.DirToVec(dc.Direction)
	offsets := aoe.Cone(fdx, fdy, depth, spread)

	var hitBuf []*ecs.Entity
	for _, off := range offsets {
		tx, ty := ox+off.X, oy+off.Y

		level.AddTileAnim(tx, ty, z, &world.TileAnim{
			Resource:   "fx",
			SpriteX:    128,
			SpriteY:    0,
			FrameCount: 4,
			FrameSpeed: 4,
			TTL:        60,
			LightLevel: 255,
		})

		hitBuf = hitBuf[:0]
		level.GetEntitiesAt(tx, ty, z, &hitBuf)
		for _, e := range hitBuf {
			if e == entity || e.HasComponent(rlcomponents.Dead) {
				continue
			}
			entityhelpers.SavingThrow(level, entity, e, saveStat, saveDC, damageType, damageDice)
		}
	}

	rlenergy.SetActionCost(entity, energy.CostAttack)
	return nil
}
