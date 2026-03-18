package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/state"
	mlge_text "github.com/mechanical-lich/mlge/text"

	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/ui"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const numLevels = 4

type MainState struct {
	level         *world.Level
	CurrentZ      int
	CameraX       int
	CameraY       int
	keys          []ebiten.Key
	pressDelay    int
	systemManager *ecs.SystemManager

	Player          *ecs.Entity
	PlayerInputHalt bool
	gm              gamemaster.GameMaster
	updateDelay     int
	gui             *ui.GUI
	pause           bool
	tick            int
	inventoryView   *InventoryView
}

func NewMainState() (*MainState, error) {
	m := MainState{systemManager: &ecs.SystemManager{}, gui: ui.NewGUI(&GUIViewMain{}), gm: gamemaster.GameMaster{}}

	pX := 50
	pY := 50
	m.level = world.NewLevel(100, 100, numLevels, world.NewDefaultTheme())

	for z := 0; z < numLevels; z++ {
		switch utility.GetRandom(0, 3) {
		case 0:
			generation.GenerateStation(m.level, z, 100, 100)
		case 1:
			generation.GenerateRoundStation(m.level, z)
		case 2:
			generation.GenerateRectangleStation(m.level, z)
		}

		m.level.Polish(z)
		m.gm.Init(m.level, z)
	}

	// Temp stair gen
	m.level.SetTileTypeAt(pX, pY, 0, world.TypeStairsUp)
	m.level.Polish(0)

	m.level.SetTileTypeAt(pX, pY, 1, world.TypeStairsDown)
	m.level.Polish(1)

	// Setup Systems
	m.systemManager.AddSystem(system.InitiativeSystem{})
	m.systemManager.AddSystem(system.StatusConditionSystem{})
	m.systemManager.AddSystem(&system.PlayerSystem{})
	m.systemManager.AddSystem(&system.AISystem{})
	m.systemManager.AddSystem(&system.LightSystem{})

	// Create player
	var err error
	m.Player, err = factory.Create("player", pX, pY)
	if err != nil {
		return nil, err
	}
	// Set Z=0 for the player
	m.Player.GetComponent("Position").(*component.PositionComponent).SetPosition(pX, pY, 0)
	m.level.AddEntity(m.Player)

	item, _ := factory.Create("health", pX+2, pY)
	if item != nil {
		item.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+2, pY, 0)
		m.level.AddEntity(item)
	}

	m.UpdateEntities()

	// Setup inventory menu
	m.inventoryView = NewInventoryView(m.Player)

	eventsystem.EventManager.RegisterListener(&m, eventsystem.Stairs)
	eventsystem.EventManager.RegisterListener(&m, eventsystem.DropItem)

	// Register queued message listener so message.MessageEvent queued via message.PostMessage
	// are flushed into message.MessageLog by our listener.
	event.GetQueuedInstance().RegisterListener(&queuedMessageListener{level: m.level, player: m.Player}, message.MessageEventType)

	return &m, nil
}

func (s *MainState) Done() bool {
	return false
}

func (s *MainState) Update() state.StateInterface {
	// First, process any queued events (including message events)
	event.GetQueuedInstance().HandleQueue()

	s.gui.Update(s)
	s.inventoryView.Update()
	s.tick++
	if !s.pause {
		if s.Player != nil && s.updateDelay <= 0 {
			playerC := s.Player.GetComponent("PlayerComponent").(*component.PlayerComponent)
			pc := s.Player.GetComponent("Position").(*component.PositionComponent)

			// The amount of ticks it takes to push the command again if the key is held down.
			if s.pressDelay > 0 {
				s.pressDelay--
			}

			// Pause the game for the player to take their turn
			if s.Player.HasComponent("MyTurn") {
				s.PlayerInputHalt = true
				// Open Inventory
				if inpututil.IsKeyJustPressed(ebiten.KeyI) {
					s.inventoryView.Visible = true
				}
				// Handle input
				s.keys = inpututil.AppendPressedKeys([]ebiten.Key{})
				for _, k := range s.keys {
					if s.pressDelay == 0 {
						playerC.PushCommand(k.String())
						s.pressDelay = config.PressDelay

						s.PlayerInputHalt = false
					}
				}
			}

			if !s.PlayerInputHalt {
				cS := system.CleanUpSystem{}
				cS.Update(s.level)
				s.UpdateEntities()
			}
			s.CameraX = pc.GetX()
			s.CameraY = pc.GetY()
			s.updateDelay = config.UpdateDelay
		}

		if s.tick%20 == 0 {
			for _, entity := range s.level.Entities {
				if entity.HasComponent("AppearanceComponent") {
					ac := entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent)
					ac.Update()
				}
			}
		}

		s.updateDelay--
	}

	// Close inventory
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && s.inventoryView.Visible {
		s.inventoryView.Visible = false
	}
	return nil
}

func (s *MainState) Draw(screen *ebiten.Image) {
	levelImage := s.level.Render(s.CameraX, s.CameraY, s.CurrentZ, config.GameWidth/config.SpriteWidth, config.GameHeight/config.SpriteHeight, false, true)
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(levelImage, op)

	s.gui.Draw(screen, s)
	s.inventoryView.Draw(screen)
}

func (s *MainState) DrawPlayerMessages(screen *ebiten.Image) {
	// Player messages
	if s.Player != nil {
		x := 0
		y := 0
		for _, v := range message.MessageLog {
			mlge_text.Draw(screen, v, 24, x, y, color.RGBA{255, 0, 0, 255})
			y += 24
		}
	}
}

func (s *MainState) UpdateEntities() {
	system.LightSystem{}.ClearLights(s.level, s.CurrentZ)

	for _, entity := range s.level.Entities {
		if entity == nil {
			fmt.Println("Entity is nil, probably picked up.")
			continue
		}
		s.systemManager.UpdateSystemsForEntity(s.level, entity)
	}
}

func (s *MainState) HandleEvent(data event.EventData) error {
	switch data.GetType() {
	case eventsystem.Stairs:
		stairsEvent := data.(eventsystem.StairsEventData)
		pc := s.Player.GetComponent("Position").(*component.PositionComponent)

		if stairsEvent.Up {
			if s.CurrentZ < numLevels-1 {
				s.CurrentZ++
				s.level.PlaceEntity(pc.GetX(), pc.GetY(), s.CurrentZ, s.Player)
			}
		} else {
			if s.CurrentZ > 0 {
				s.CurrentZ--
				s.level.PlaceEntity(pc.GetX(), pc.GetY(), s.CurrentZ, s.Player)
			}
		}

		s.UpdateEntities()
	case eventsystem.DropItem:
		dropItemEvent := data.(eventsystem.DropItemEventData)
		dropItemEvent.Item.GetComponent("Position").(*component.PositionComponent).SetPosition(dropItemEvent.X, dropItemEvent.Y, dropItemEvent.Z)
		s.level.AddEntity(dropItemEvent.Item)
	}

	return nil
}
