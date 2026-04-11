package action

import (
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// SpreadOvergrowthAction marks the entity's current tile as overgrown and
// attempts one CA step of overgrowth spread to neighboring tiles within radius.
// Overgrowth crosses Z-levels through stair tiles.
type SpreadOvergrowthAction struct {
	Params ActionParams
}

func (a SpreadOvergrowthAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	return a.Params.Cost(energy.CostMove)
}

func (a SpreadOvergrowthAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	return entity.HasComponent(component.Position)
}

func (a SpreadOvergrowthAction) Execute(entity *ecs.Entity, level *world.Level) error {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	ox, oy, oz := pc.GetX(), pc.GetY(), pc.GetZ()

	radius := a.Params.Radius
	if radius <= 0 {
		radius = 4
	}
	spreadChance := a.Params.SpreadChance
	if spreadChance <= 0 {
		spreadChance = 40
	}

	// Always seed the entity's current tile.
	if t := level.GetTilePtr(ox, oy, oz); t != nil && !t.IsAir() {
		level.SetOvergrown(ox, oy, oz, true)
	}

	// Collect all overgrown tiles within the radius, including adjacent Z-levels.
	type coord struct{ x, y, z int }
	var sources []coord

	r2 := radius * radius
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			for _, dz := range []int{-1, 0, 1} {
				if level.IsOvergrown(ox+dx, oy+dy, oz+dz) {
					sources = append(sources, coord{ox + dx, oy + dy, oz + dz})
				}
			}
		}
	}

	cardinals := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, src := range sources {
		// Spread to cardinal neighbors on the same Z-level.
		for _, d := range cardinals {
			nx, ny := src.x+d[0], src.y+d[1]
			ddx, ddy := nx-ox, ny-oy
			if ddx*ddx+ddy*ddy > r2 {
				continue
			}
			if rand.Intn(100) >= spreadChance {
				continue
			}
			if t := level.GetTilePtr(nx, ny, src.z); t != nil && !level.IsOvergrown(nx, ny, src.z) && !t.IsAir() {
				level.SetOvergrown(nx, ny, src.z, true)
			}
		}

		// Spread through stairs to adjacent Z-levels.
		srcTile := level.GetTilePtr(src.x, src.y, src.z)
		if srcTile == nil {
			continue
		}
		def := rlworld.TileDefinitions[srcTile.Type]
		if def.StairsUp {
			if rand.Intn(100) < spreadChance {
				if t := level.GetTilePtr(src.x, src.y, src.z+1); t != nil && !level.IsOvergrown(src.x, src.y, src.z+1) && !t.IsAir() {
					level.SetOvergrown(src.x, src.y, src.z+1, true)
				}
			}
		}
		if def.StairsDown {
			if rand.Intn(100) < spreadChance {
				if t := level.GetTilePtr(src.x, src.y, src.z-1); t != nil && !level.IsOvergrown(src.x, src.y, src.z-1) && !t.IsAir() {
					level.SetOvergrown(src.x, src.y, src.z-1, true)
				}
			}
		}
	}

	rlenergy.SetActionCost(entity, a.Cost(entity, level))
	return nil
}
