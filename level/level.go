package level

import (
	"errors"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mechanical-lich/game-engine/ecs"

	"github.com/mechanical-lich/game-engine/resource"
	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/utility"
)

type Level struct {
	data          [][]Tile
	Width, Height int
	Theme         Theme
	Entities      []*ecs.Entity
}

func NewLevel(width int, height int, theme Theme) (level *Level) {
	level = &Level{Width: width, Height: height, Theme: theme}

	data := make([][]Tile, width, height)
	for x := 0; x < width; x++ {
		col := []Tile{}
		for y := 0; y < height; y++ {
			col = append(col, Tile{Type: Type_Open, TileIndex: theme.Open[0], X: x, Y: y, level: level, ForgroundColor: color.White, BackgroundColor: color.Black})
		}
		data[x] = append(data[x], col...)
	}

	level.data = data
	return
}

// Tile utility functions
func (level *Level) GetTileAt(x int, y int) (tile *Tile) {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		tile = &level.data[x][y]
	}
	return
}

func (level *Level) IsTileSolid(x int, y int) bool {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		return level.data[x][y].Solid
	}
	return false
}

func (level *Level) GetTileType(x int, y int) TileType {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		return level.data[x][y].Type
	}
	return -1
}

func (level *Level) SetTileType(x int, y int, t TileType) error {
	tile := level.GetTileAt(x, y)
	if tile == nil {
		return errors.New("invalid tile")
	}

	tile.Type = t

	return nil
}

func (level *Level) TileNeighbors(x, y int, nType TileType) bool {
	for _, offset := range [][]int{
		{-1, 0},
		{1, 0},
		{0, -1},
		{0, 1},
	} {
		if n := level.GetTileType(x+offset[0], y+offset[1]); n == nType {
			return true
		}
	}

	return false
}

func (level *Level) Polish() {
	// Build tunnel walls
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			tile := level.GetTileAt(x, y)
			if tile != nil {
				if tile.Type == Type_Open {
					if level.TileNeighbors(x, y, Type_MaintenanceTunnelFLoor) {
						level.SetTileType(x, y, Type_MaintenanceTunnelWall)
					}
				}
			}
		}
	}

	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			tile := level.GetTileAt(x, y)
			if tile != nil {

				// Pick a random variant based on type.
				// TODO there has to be a less hard coded way of doing this.
				switch tile.Type {
				case Type_Wall:
					tile.TileIndex = level.Theme.Wall[utility.GetRandom(0, len(level.Theme.Wall))]
					tile.Solid = true
					belowTile := level.GetTileAt(x, y+1)
					if belowTile != nil {
						if belowTile.Type == Type_Wall || belowTile.Type == Type_Door || belowTile.Type == Type_MaintenanceTunnelDoor {
							tile.TileIndex = level.Theme.WallTop[utility.GetRandom(0, len(level.Theme.WallTop))]
						}
					}
					tile.ForgroundColor = level.Theme.ForgroundColor
					tile.BackgroundColor = level.Theme.BackgroundColor
				case Type_Floor:
					tile.TileIndex = level.Theme.Floor[utility.GetRandom(0, len(level.Theme.Floor))]
					tile.ForgroundColor = level.Theme.ForgroundColor
					tile.BackgroundColor = level.Theme.BackgroundColor
				case Type_MaintenanceTunnelWall:
					tile.TileIndex = level.Theme.MaintenanceTunnelWall[utility.GetRandom(0, len(level.Theme.MaintenanceTunnelWall))]
					tile.Solid = true
					belowTile := level.GetTileAt(x, y+1)
					if belowTile != nil {
						if belowTile.Type == Type_MaintenanceTunnelWall || belowTile.Type == Type_MaintenanceTunnelDoor || belowTile.Type == Type_Door {
							tile.TileIndex = level.Theme.MaintenanceTunnelTop[utility.GetRandom(0, len(level.Theme.MaintenanceTunnelTop))]
						}
					}
					tile.ForgroundColor = level.Theme.SecondaryForgroundColor
					tile.BackgroundColor = level.Theme.SecondaryBackgroundColor
				case Type_MaintenanceTunnelFLoor:
					tile.TileIndex = level.Theme.MaintenanceTunnelFloor[utility.GetRandom(0, len(level.Theme.MaintenanceTunnelFloor))]
					tile.ForgroundColor = level.Theme.SecondaryForgroundColor
					tile.BackgroundColor = level.Theme.SecondaryBackgroundColor
				case Type_Stairs:
					tile.TileIndex = level.Theme.Stairs[utility.GetRandom(0, len(level.Theme.Stairs))]
					tile.ForgroundColor = level.Theme.SecondaryForgroundColor
					tile.BackgroundColor = level.Theme.SecondaryBackgroundColor
				case Type_Open:
					tile.TileIndex = level.Theme.Open[utility.GetRandom(0, len(level.Theme.Open))]
					tile.ForgroundColor = level.Theme.OpenForgroundColor
					tile.BackgroundColor = level.Theme.OpenBackgroundColor
				case Type_Door:
					tile.TileIndex = level.Theme.Door[utility.GetRandom(0, len(level.Theme.Door))]
					//tile.Solid = true
					tile.ForgroundColor = level.Theme.SecondaryForgroundColor
					tile.BackgroundColor = level.Theme.SecondaryBackgroundColor

				case Type_MaintenanceTunnelDoor:
					tile.TileIndex = level.Theme.MaintenanceTunnelDoor[utility.GetRandom(0, len(level.Theme.MaintenanceTunnelDoor))]
					//tile.Solid = true
					tile.ForgroundColor = level.Theme.SecondaryForgroundColor
					tile.BackgroundColor = level.Theme.SecondaryBackgroundColor
				}

			}
		}
	}

}

// Entity functions
func (level *Level) PlaceEntity(x int, y int, entity *ecs.Entity) {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		tile := &level.data[x][y]
		pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
		oldTile := &level.data[pc.GetX()][pc.GetY()]
		for i := 0; i < len(oldTile.Entities); i++ {
			if oldTile.Entities[i] == entity {
				oldTile.Entities = append(oldTile.Entities[:i], oldTile.Entities[i+1:]...)
			}
		}
		tile.Entities = append(tile.Entities, entity)
		pc.SetPosition(x, y)
	}
}

func (level *Level) GetEntityAt(x int, y int) (entity *ecs.Entity) {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		tile := &level.data[x][y]
		if len(tile.Entities) > 0 {
			return tile.Entities[0]
		}
	}

	entity = nil
	return
}
func (level *Level) GetEntitiesAround(x int, y int, width int, height int) (entities []*ecs.Entity) {
	left := x - width/2
	right := x + width/2
	up := y - height/2
	down := y + height/2

	for x := left; x < right; x++ {
		for y := up; y < down; y++ {
			tile := level.GetTileAt(x, y)
			if tile != nil {
				if len(tile.Entities) > 0 {
					entity := tile.Entities[0]

					if entity.HasComponent("PositionComponent") {
						pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
						if pc.GetX() >= left && pc.GetX() <= right && pc.GetY() >= up && pc.GetY() <= down {
							entities = append(entities, entity)
						}
					}
				}
			}
		}
	}
	return
}
func (level *Level) GetSolidEntityAt(x int, y int) (entity *ecs.Entity) {
	if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
		tile := &level.data[x][y]
		if len(tile.Entities) > 0 {
			if tile.Entities[0].HasComponent("SolidComponent") {

				return tile.Entities[0]
			}
		}
	}

	entity = nil
	return
}

func (level *Level) RemoveEntity(entity *ecs.Entity) {
	if entity.HasComponent("PositionComponent") {
		pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
		x := pc.GetX()
		y := pc.GetY()

		if x < level.Width && y < level.Height && x >= 0 && y >= 0 {
			tile := &level.data[pc.GetX()][pc.GetY()]
			for i := 0; i < len(tile.Entities); i++ {
				if tile.Entities[i] == entity {
					tile.Entities = append(tile.Entities[:i], tile.Entities[i+1:]...)
				}
			}
		}
	}
	for i := 0; i < len(level.Entities); i++ {
		if level.Entities[i] == entity {
			level.Entities = append(level.Entities[:i], level.Entities[i+1:]...)
			return
		}
	}
}

func (level *Level) AddEntity(entity *ecs.Entity) {
	level.Entities = append(level.Entities, entity)
	if entity.HasComponent("PositionComponent") {
		pc := entity.GetComponent("PositionComponent").(*component.PositionComponent)
		x := pc.GetX()
		y := pc.GetY()
		level.PlaceEntity(x, y, entity)
	}
}

func (level *Level) Render(aX int, aY int, width int, height int, blind bool, centered bool, useLos bool, useFog bool) *ebiten.Image {
	output := ebiten.NewImage(width*config.SpriteWidth, height*config.SpriteHeight)
	left := aX - width/2
	right := aX + width/2
	up := aY - height/2
	down := aY + height/2

	if !centered {
		left = aX
		right = aX + width - 1
		up = aY
		down = aY + height
	}

	screenX := 0
	screenY := 0
	for x := left; x <= right; x++ {
		screenY = 0
		for y := up; y <= down; y++ {
			tile := level.GetTileAt(x, y)

			if blind {
				if y < aY-height/4 || y > aY+height/4 || x > aX+width/4 || x < aX-width/4 {
					tile = nil
				}
			}

			//Draw tile
			tX := float64(screenX * config.SpriteWidth)
			tY := float64(screenY * config.SpriteHeight)

			// // Figure out colors
			backgroundColor := level.Theme.BackgroundColor
			forgroundColor := level.Theme.ForgroundColor

			seen := true

			if tile != nil {
				// LOS logic
				if useLos {
					seen = los(aX, aY, tile.X, tile.Y, level)
				}
				forgroundColor = tile.ForgroundColor
				backgroundColor = tile.BackgroundColor
			}

			////Draw background square
			ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, backgroundColor)

			// Draw forground
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tX, tY)

			// Set color
			op.ColorM.ScaleWithColor(forgroundColor)

			// Default Tile
			if tile != nil {
				// Real tile
				sX := tile.TileIndex * config.SpriteWidth
				output.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)

				if !seen {
					if tile.Seen {
						ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, color.RGBA{0, 0, 0, 150})
					} else {
						ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, level.Theme.OpenBackgroundColor)
					}

				} else {
					tile.Seen = true

					//Draw entity on tile.  We do this here to prevent yet another loop. ;)
					entity := level.GetEntityAt(x, y)
					if entity != nil && seen {
						level.DrawEntity(output, entity, tX, tY)
					}
					// Draw fog
					if useFog {
						dist := utility.Distance(aX, aY, tile.X, tile.Y)
						dist = 255 * dist / 20
						if dist > 255 {
							dist = 255
						}
						fogColor := color.RGBA{0, 0, 0, uint8(dist)}

						ebitenutil.DrawRect(output, tX, tY, config.SpriteWidth, config.SpriteHeight, fogColor)
					}
				}
			}

			screenY++
		}
		screenX++
	}

	return output
}

func (level *Level) DrawEntity(screen *ebiten.Image, entity *ecs.Entity, x float64, y float64) {
	//Draw entity on tile.
	if entity != nil {
		if entity.HasComponent("AppearanceComponent") {
			ac := entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent)
			// dir := 0
			// if entity.HasComponent("DirectionComponent") {
			// 	dc := entity.GetComponent("DirectionComponent").(*component.DirectionComponent)
			// 	dir = dc.Direction
			// }

			op := &ebiten.DrawImageOptions{}

			// Dead transformation
			if entity.HasComponent("DeadComponent") {
				op.GeoM.Scale(1, -1)
				op.GeoM.Translate(0, float64(config.SpriteHeight))
			}

			//Position
			op.GeoM.Translate(x, y)
			//Color
			//op.ColorM.ScaleWithColor(color.RGBA{ac.R, ac.G, ac.B, 255})
			// TODO - I don't like this.  The appearance component should specify the resource.
			screen.DrawImage(resource.Textures["entities"].SubImage(image.Rect(ac.SpriteX, ac.SpriteY, ac.SpriteX+config.SpriteWidth, ac.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)

			//Draw FX
			if entity.HasComponent("AttackComponent") {
				attackC := entity.GetComponent("AttackComponent").(*component.AttackComponent)
				if attackC.Frame == 3 {
					entity.RemoveComponent("AttackComponent")
				} else {
					xOffset := attackC.SpriteX + (attackC.Frame * config.SpriteWidth)
					op := &ebiten.DrawImageOptions{}
					//op.GeoM.Scale(float64(config.TileWidth/config.SpriteWidth), float64(config.TileHeight/config.SpriteHeight))
					op.GeoM.Translate(x, y)
					screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+config.SpriteWidth, attackC.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)
					attackC.Frame++
				}
			}
		}
	}
}

func los(pX int, pY int, tX int, tY int, level *Level) bool {
	deltaX := pX - tX
	deltaY := pY - tY

	absDeltaX := math.Abs(float64(deltaX))
	absDeltaY := math.Abs(float64(deltaY))

	signX := utility.Sgn(deltaX)
	signY := utility.Sgn(deltaY)

	if absDeltaX > absDeltaY {
		t := absDeltaY*2 - absDeltaX
		for {
			if t >= 0 {
				tY += signY
				t -= absDeltaX * 2
			}

			tX += signX
			t += absDeltaY * 2

			if tX == pX && tY == pY {
				return true
			}
			if level.IsTileSolid(tX, tY) || level.GetTileType(tX, tY) == Type_Door || level.GetTileType(tX, tY) == Type_MaintenanceTunnelDoor {
				break
			}
		}
		return false
	}

	t := absDeltaX*2 - absDeltaY

	for {
		if t >= 0 {
			tX += signX
			t -= absDeltaY * 2
		}
		tY += signY
		t += absDeltaX * 2
		if tX == pX && tY == pY {
			return true
		}

		if level.IsTileSolid(tX, tY) || level.GetTileType(tX, tY) == Type_Door || level.GetTileType(tX, tY) == Type_MaintenanceTunnelDoor {
			break
		}
	}

	return false

}
