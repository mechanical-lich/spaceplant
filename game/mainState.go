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
	"github.com/mechanical-lich/spaceplant/utility"
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

	Player          *ecs.Entity
	PlayerInputHalt bool
	gm              GameMaster
	updateDelay     int
}

func NewMainState() (*MainState, error) {
	m := MainState{systemManager: &ecs.SystemManager{}}

	pX := 50
	pY := 50
	for i := 0; i < numLevels; i++ {
		m.levels = append(m.levels, level.NewLevel(100, 100, level.NewDefaultTheme()))

		switch utility.GetRandom(0, 3) {
		case 0:
			generation.GenerateStation(m.levels[i], 100, 100)
		case 1:
			generation.GenerateRoundStation(m.levels[i])
		case 2:
			generation.GenerateRectangleStation(m.levels[i])

		}

		//generation.GenerateStation(m.levels[i], 100, 100)
		//generation.GenerateRoundStation(m.levels[i])
		//generation.GenerateRectangleStation(m.levels[i])
		m.levels[i].Polish()
	}

	// Load Blueprints
	err := factory.FactoryLoad("entities.blueprints")
	if err != nil {
		return nil, err
	}
	//TODO feed gm the current level
	m.gm = GameMaster{}
	m.gm.Init(m.levels[0])
	// Setup Systems
	m.systemManager.AddSystem(system.InitiativeSystem{})
	m.systemManager.AddSystem(system.StatusConditionSystem{})
	m.systemManager.AddSystem(&system.PlayerSystem{})
	m.systemManager.AddSystem(&system.AISystem{})
	m.systemManager.AddSystem(&system.LightSystem{})

	// Create player
	// TODO - This shouldn't be permenant
	m.Player, err = factory.Create("player", pX, pY)
	if err != nil {
		return nil, err
	}

	m.levels[m.CurrentLevel].AddEntity(m.Player)

	m.UpdateEntities()

	return &m, nil
}

func (s *MainState) Update() {
	if s.Player != nil && s.updateDelay <= 0 {
		playerC := s.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
		pc := s.Player.GetComponent("PositionComponent").(*component.PositionComponent)

		// The amount of ticks it takes to push the command again if the key is held down.
		if s.pressDelay > 0 {
			s.pressDelay--
		}

		// Pause the game for the player to take their turn
		if s.Player.HasComponent("MyTurnComponent") {
			s.PlayerInputHalt = true

			// Handle input
			s.keys = inpututil.AppendPressedKeys([]ebiten.Key{})
			for _, k := range s.keys {
				if s.pressDelay == 0 {
					playerC.PushCommand(k.String())
					s.pressDelay = config.PressDelay

					s.PlayerInputHalt = false
					//s.gm.Update(pc.GetX(), pc.GetY())
				}
			}
		}

		if !s.PlayerInputHalt {
			cS := system.CleanUpSystem{}
			cS.Update(s.levels[s.CurrentLevel])
			s.UpdateEntities()

		}
		s.CameraX = pc.GetX()
		s.CameraY = pc.GetY()
		s.updateDelay = config.UpdateDelay
	}
	s.updateDelay--
}

func (s *MainState) Draw(screen *ebiten.Image) {
	levelImage := s.levels[s.CurrentLevel].Render(s.CameraX, s.CameraY, config.ScreenWidth/config.SpriteWidth, config.ScreenHeight/config.SpriteHeight, false, true)
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

func (s *MainState) UpdateEntities() {
	system.LightSystem{}.ClearLights(s.levels[s.CurrentLevel])

	for _, entity := range s.levels[s.CurrentLevel].Entities {
		s.systemManager.UpdateSystemsForEntity(s.levels[s.CurrentLevel], entity)
	}
}
