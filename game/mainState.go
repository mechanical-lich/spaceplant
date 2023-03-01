package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	text_ext "github.com/mechanical-lich/game-engine/text"

	"github.com/mechanical-lich/game-engine/ecs"

	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/generation"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/system"
)

const numLevels = 4

type MainState struct {
	levels        []*level.Level
	CurrentLevel  int
	CameraX       int
	CameraY       int
	keys          []ebiten.Key
	pressDelay    int
	systemManager *ecs.SystemManager

	Player *ecs.Entity

	gm GameMaster
}

func NewMainState() (*MainState, error) {
	m := MainState{systemManager: &ecs.SystemManager{}}

	//generation.CarveRoom(m.level, 10, 10, 20, 20, level.Type_Wall, level.Type_Floor, false)
	//fmt.Println(generation.RoomIntersects(m.level, 10, 6, 5, 5))
	// generation.CarveRoom(m.level, 11, 2, 10, 3, level.Type_MaintenanceTunnelWall, level.Type_MaintenanceTunnelFLoor, true)

	// generation.CarveRoom(m.level, 2, 11, 30, 10, level.Type_Wall, level.Type_Floor, true)

	// m.level.SetTileType(5, 11, level.Type_Door)
	// m.level.SetTileType(11, 3, level.Type_MaintenanceTunnelDoor)

	for i := 0; i < numLevels; i++ {
		m.levels = append(m.levels, level.NewLevel(100, 100, level.NewDefaultTheme()))
		//generation.GenerateStation(m.levels[i], 100, 100)
		//generation.GenerateRoundStation(m.levels[i])
		generation.GenerateRectangleStation(m.levels[i])
		m.levels[i].Polish()
	}

	// Load Blueprints
	err := factory.FactoryLoad("entities.blueprints")
	if err != nil {
		return nil, err
	}
	//TODO feed gm the current level
	m.gm = GameMaster{}
	//m.gm.Init(m.levels[0])
	// Setup Systems
	m.systemManager.AddSystem(system.InitiativeSystem{})
	m.systemManager.AddSystem(system.StatusConditionSystem{})
	m.systemManager.AddSystem(&system.PlayerSystem{})
	m.systemManager.AddSystem(&system.AISystem{})

	// Create player
	// TODO - This shouldn't be permenant
	m.Player, err = factory.Create("player", 2, 2)
	if err != nil {
		return nil, err
	}

	m.levels[m.CurrentLevel].AddEntity(m.Player)

	// e, err := factory.Create("viner", 48, 50)
	// if err != nil {
	// 	return nil, err
	// }

	// m.levels[m.CurrentLevel].AddEntity(e)

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
		playerC := s.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
		pc := s.Player.GetComponent("PositionComponent").(*component.PositionComponent)

		// The amount of ticks it takes to push the command again if the key is held down.
		if s.pressDelay > 0 {
			s.pressDelay--
		}

		s.keys = inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range s.keys {
			if s.pressDelay == 0 {
				playerC.PushCommand(k.String())
				for _, entity := range s.levels[s.CurrentLevel].Entities {
					s.systemManager.UpdateSystemsForEntity(s.levels[s.CurrentLevel], entity)
				}
				s.pressDelay = config.PressDelay
				//s.gm.Update(pc.GetX(), pc.GetY())

			}
		}

		s.CameraX = pc.GetX()
		s.CameraY = pc.GetY()
	}

	cS := system.CleanUpSystem{}
	cS.Update(s.levels[s.CurrentLevel])

}

func (s *MainState) Draw(screen *ebiten.Image) {
	//s.DrawLevel(screen, s.CameraX, s.CameraY, 10, 10, false, true)
	levelImage := s.levels[s.CurrentLevel].Render(s.CameraX, s.CameraY, config.ScreenWidth/config.SpriteWidth, config.ScreenHeight/config.SpriteHeight, false, true, config.Los, config.Fog)
	op := &ebiten.DrawImageOptions{}
	//op.GeoM.Scale(1.5, 1.5)
	screen.DrawImage(levelImage, op)

}

func (s *MainState) DrawPlayerMessages(screen *ebiten.Image) {
	// Player messages
	if s.Player != nil {
		playerC := s.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
		x := 0
		y := 0
		for _, v := range playerC.MessageLog {
			text.Draw(screen, v, text_ext.MplusNormalFont, x, y, color.RGBA{255, 0, 0, 255})
			y += 24
		}
	}
}
