package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/game-engine/entity"
	"github.com/mechanical-lich/game-engine/system"

	"github.com/mechanical-lich/spaceplant/components"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/generation"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/systems"
)

type MainState struct {
	level *level.Level

	CameraX       int
	CameraY       int
	keys          []ebiten.Key
	systemManager *system.SystemManager

	Player *entity.Entity
}

func NewMainState() (*MainState, error) {
	m := MainState{systemManager: &system.SystemManager{}}
	m.level = level.NewLevel(100, 100, level.NewDefaultTheme())

	//generation.CarveRoom(m.level, 10, 10, 20, 20, level.Type_Wall, level.Type_Floor, false)
	//fmt.Println(generation.RoomIntersects(m.level, 10, 6, 5, 5))
	// generation.CarveRoom(m.level, 11, 2, 10, 3, level.Type_MaintenanceTunnelWall, level.Type_MaintenanceTunnelFLoor, true)

	// generation.CarveRoom(m.level, 2, 11, 30, 10, level.Type_Wall, level.Type_Floor, true)

	// m.level.SetTileType(5, 11, level.Type_Door)
	// m.level.SetTileType(11, 3, level.Type_MaintenanceTunnelDoor)
	generation.GenerateStation(m.level, 100, 100)
	m.level.Polish()

	// Setup Systems
	m.systemManager.AddSystem(systems.InitiativeSystem{})
	m.systemManager.AddSystem(systems.StatusConditionSystem{})
	m.systemManager.AddSystem(&systems.PlayerSystem{})
	m.systemManager.AddSystem(&systems.AISystem{})

	// Load Blueprints
	err := factory.FactoryLoad("entities.blueprints")
	if err != nil {
		return nil, err
	}

	// Create player
	// TODO - This shouldn't be permenant
	m.Player, err = factory.Create("player", 50, 50)
	if err != nil {
		return nil, err
	}

	m.level.AddEntity(m.Player)

	e, err := factory.Create("creeper", 48, 50)
	if err != nil {
		return nil, err
	}

	m.level.AddEntity(e)

	// e, err = factory.Create("crewmember", 7, 8)
	// if err != nil {
	// 	return nil, err
	// }

	// m.level.AddEntity(e)

	// e, err = factory.Create("officer", 7, 7)
	// if err != nil {
	// 	return nil, err
	// }

	// m.level.AddEntity(e)

	return &m, nil
}

func (s *MainState) Update() {
	if s.Player != nil {
		playerC := s.Player.GetComponent("PlayerComponent").(*components.PlayerComponent)

		s.keys = inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range s.keys {
			if inpututil.IsKeyJustPressed(k) {
				playerC.PushCommand(k.String())
				for _, entity := range s.level.Entities {
					s.systemManager.UpdateSystemsForEntity(s.level, entity)
				}
			}
		}

		pc := s.Player.GetComponent("PositionComponent").(*components.PositionComponent)
		s.CameraX = pc.GetX()
		s.CameraY = pc.GetY()
	}

	cS := systems.CleanUpSystem{}
	cS.Update(s.level)

}

func (s *MainState) Draw(screen *ebiten.Image) {
	//s.DrawLevel(screen, s.CameraX, s.CameraY, 10, 10, false, true)
	levelImage := s.level.Render(s.CameraX, s.CameraY, config.ScreenWidth/config.SpriteWidth, config.ScreenHeight/config.SpriteHeight, false, true, false)
	op := &ebiten.DrawImageOptions{}
	//op.GeoM.Scale(1.5, 1.5)
	screen.DrawImage(levelImage, op)
}

// func (s *MainState) DrawLevel(screen *ebiten.Image, aX int, aY int, width int, height int, blind bool, centered bool) {
// 	left := aX - width/2
// 	right := aX + width/2
// 	up := aY - height/2
// 	down := aY + height/2

// 	if !centered {
// 		left = aX
// 		right = aX + width - 1
// 		up = aY
// 		down = aY + height
// 	}

// 	screenX := 0
// 	screenY := 0
// 	for x := left; x <= right; x++ {
// 		screenY = 0
// 		for y := up; y <= down; y++ {
// 			tile := s.level.GetTileAt(x, y)
// 			if blind {
// 				if y < aY-height/4 || y > aY+height/4 || x > aX+width/4 || x < aX-width/4 {
// 					tile = nil
// 				}
// 			}

// 			//Draw tile
// 			tX := float64(screenX * config.TileWidth)
// 			tY := float64(screenY * config.TileHeight)

// 			// Figure out colors
// 			backgroundColor := s.level.Theme.BackgroundColor
// 			forgroundColor := s.level.Theme.ForgroundColor

// 			if tile != nil {
// 				if tile.Type == level.Type_Open {
// 					backgroundColor = s.level.Theme.OpenBackgroundColor
// 				}

// 				if tile.Type == level.Type_Open {
// 					forgroundColor = s.level.Theme.OpenForgroundColor
// 				}
// 			}

// 			//Draw background square
// 			ebitenutil.DrawRect(screen, tX, tY, config.TileWidth, config.TileHeight, backgroundColor)

// 			// Draw forground
// 			op := &ebiten.DrawImageOptions{}
// 			op.GeoM.Scale(float64(config.TileWidth/config.SpriteWidth), float64(config.TileHeight/config.SpriteHeight))
// 			op.GeoM.Translate(tX, tY)
// 			//	op.ColorM.Scale(0, 0, 0, 1)

// 			// Set color
// 			op.ColorM.ScaleWithColor(forgroundColor)

// 			if tile == nil {
// 				sX := s.level.Theme.Open[0] * config.SpriteWidth
// 				screen.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)
// 				continue
// 			} else {
// 				sX := tile.TileIndex * config.SpriteWidth
// 				screen.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)
// 			}

// 			//Draw entity on tile.  We do this here to prevent yet another loop. ;)
// 			entity := s.level.GetEntityAt(x, y)
// 			if entity != nil {
// 				s.DrawEntity(screen, entity, tX, tY)
// 			}

// 			screenY++
// 		}
// 		screenX++
// 	}
// }

// func (s *MainState) DrawEntity(screen *ebiten.Image, entity *entity.Entity, x float64, y float64) {
// 	//Draw entity on tile.
// 	if entity != nil {
// 		if entity.HasComponent("AppearanceComponent") {
// 			ac := entity.GetComponent("AppearanceComponent").(*components.AppearanceComponent)
// 			dir := 0
// 			if entity.HasComponent("DirectionComponent") {
// 				dc := entity.GetComponent("DirectionComponent").(*components.DirectionComponent)
// 				dir = dc.Direction
// 			}

// 			op := &ebiten.DrawImageOptions{}

// 			op.GeoM.Scale(float64(config.TileWidth/config.SpriteWidth), float64(config.TileHeight/config.SpriteHeight))
// 			if entity.HasComponent("DeadComponent") {
// 				op.GeoM.Scale(1, -1)
// 				op.GeoM.Translate(0, float64(config.TileHeight))
// 			}
// 			op.GeoM.Translate(x, y)

// 			// TODO - I don't like this.  The appearance component should specify the resource.

// 			screen.DrawImage(resource.Textures["entities"].SubImage(image.Rect(ac.SpriteX, ac.SpriteY, ac.SpriteX+config.SpriteWidth+dir*config.SpriteWidth, ac.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)

// 			//Draw FX
// 			if entity.HasComponent("AttackComponent") {
// 				attackC := entity.GetComponent("AttackComponent").(*components.AttackComponent)
// 				if attackC.Frame == 3 {
// 					entity.RemoveComponent("AttackComponent")
// 				} else {
// 					xOffset := attackC.SpriteX + (attackC.Frame * config.SpriteWidth)
// 					op := &ebiten.DrawImageOptions{}
// 					op.GeoM.Scale(float64(config.TileWidth/config.SpriteWidth), float64(config.TileHeight/config.SpriteHeight))
// 					op.GeoM.Translate(x, y)
// 					screen.DrawImage(resource.Textures["fx"].SubImage(image.Rect(xOffset, attackC.SpriteY, xOffset+config.SpriteWidth, attackC.SpriteY+config.SpriteHeight)).(*ebiten.Image), op)
// 					attackC.Frame++
// 				}
// 			}
// 		}
// 	}
// }
