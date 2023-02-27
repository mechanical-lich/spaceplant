package level

import (
	"image/color"
	"math"

	"github.com/beefsack/go-astar"
	"github.com/mechanical-lich/game-engine/ecs"
)

type TileType int

const (
	Type_Open TileType = iota
	Type_Wall
	Type_Floor
	Type_Door
	Type_Stairs
	Type_MaintenanceTunnelWall
	Type_MaintenanceTunnelFLoor
	Type_MaintenanceTunnelDoor
)

type Tile struct {
	X, Y            int
	level           *Level
	Type            TileType
	Solid           bool
	Elevation       int
	TileIndex       int // Some tile types of multiple variants for rendering
	Entities        []*ecs.Entity
	NoBudding       bool
	ForgroundColor  color.Color
	BackgroundColor color.Color
	Seen            bool
}

func (t *Tile) PathNeighbors() []astar.Pather {
	neighbors := []astar.Pather{}
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		if n := t.level.GetTileAt(t.X+offset[0], t.Y+offset[1]); n != nil &&
			!n.Solid {
			neighbors = append(neighbors, n)
		}
	}
	return neighbors
}

func (t *Tile) PathNeighborCost(to astar.Pather) float64 {
	toTile, ok := to.(*Tile)
	if !ok {
		return 100
	}
	if toTile == nil {
		return 100
	}
	if t == nil {
		return 100
	}
	cost := 10.0
	if toTile.Solid {
		cost = 100
	}

	if toTile.level.GetSolidEntityAt(toTile.X, toTile.Y) != nil {
		cost = 100
	}

	if toTile.Type == Type_Open {
		cost = 0
	}

	return cost
}

func (t *Tile) PathEstimatedCost(to astar.Pather) float64 {
	if to == nil {
		return -1
	}
	t2, ok := to.(*Tile)
	if !ok {
		return -1
	}
	if t2 == nil {
		return -1
	}

	if t == nil {
		return -1
	}

	first := math.Pow(float64(t2.X-t.X), 2)
	second := math.Pow(float64(t2.Y-t.Y), 2)
	return first + second
}
