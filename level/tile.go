package level

import (
	"image"
	"image/color"
	"math"

	"github.com/beefsack/go-astar"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/game-engine/resource"
	"github.com/mechanical-lich/spaceplant/config"
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
	Light           int
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

func (t *Tile) Draw(output *ebiten.Image, screenX, screenY int, seen bool) {
	tX := float64(screenX * config.SpriteWidth)
	tY := float64(screenY * config.SpriteHeight)
	forgroundColor := t.ForgroundColor
	backgroundColor := t.BackgroundColor

	////Draw background square
	ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, backgroundColor)

	// Draw forground
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tX, tY)

	// Set color
	if config.ColorShading {
		op.ColorM.ScaleWithColor(forgroundColor)
	}

	// Real tile
	sX := t.TileIndex * config.SpriteWidth
	output.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)

	if !seen {
		if t.Seen {
			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, color.RGBA{0, 0, 0, 220})
		} else {
			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, t.level.Theme.OpenBackgroundColor)
		}

	} else {
		t.Seen = true

		//Draw entity on tile.  We do this here to prevent yet another loop. ;)
		entity := t.level.GetEntityAt(t.X, t.Y)
		if entity != nil && seen {
			t.level.DrawEntity(output, entity, tX, tY)
		}

		// Draw fog
		if config.Lighting {
			//dist := utility.Distance(aX, aY, tile.X, tile.Y)
			//brightness := 255 * dist / 20
			// brightness := 255 - tile.Light
			// if brightness > 255 {
			// 	brightness = 255
			// }
			fogColor := color.RGBA{0, 0, 0, uint8(t.Light)}

			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, fogColor)
		}
	}
}
