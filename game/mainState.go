package game

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/mechanical-lich/game-engine/event"
	"github.com/mechanical-lich/game-engine/resource"
	"github.com/mechanical-lich/game-engine/state"
	text_ext "github.com/mechanical-lich/game-engine/text"
	"github.com/mechanical-lich/game-engine/ui"

	"github.com/mechanical-lich/game-engine/ecs"

	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/eventsystem"
	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/gamemaster"
	"github.com/mechanical-lich/spaceplant/generation"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/message"
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
		m.gm.Init(m.levels[i])
	}

	// Temp stair gen
	m.levels[0].SetTileType(pX, pY, level.Type_Stairs_Up)
	m.levels[0].Polish()

	m.levels[1].SetTileType(pX, pY, level.Type_Stairs_Down)
	m.levels[1].Polish()

	// Setup Systems
	m.systemManager.AddSystem(system.InitiativeSystem{})
	m.systemManager.AddSystem(system.StatusConditionSystem{})
	m.systemManager.AddSystem(&system.PlayerSystem{})
	m.systemManager.AddSystem(&system.AISystem{})
	m.systemManager.AddSystem(&system.LightSystem{})

	// Create player
	// TODO - This shouldn't be permenant
	var err error
	m.Player, err = factory.Create("player", pX, pY)
	if err != nil {
		return nil, err
	}
	m.levels[m.CurrentLevel].AddEntity(m.Player)
	//system.ImportInventory("laser_trimmers,helmet", m.Player)

	item, _ := factory.Create("health", pX+2, pY)
	m.levels[m.CurrentLevel].AddEntity(item)

	// item, _ = factory.Create("laser_trimmers", pX+2, pY+2)
	// m.levels[m.CurrentLevel].AddEntity(item)

	// item, _ = factory.Create("helmet", pX, pY+2)
	// m.levels[m.CurrentLevel].AddEntity(item)

	m.UpdateEntities()

	// Setup inventory menu
	m.inventoryView = NewInventoryView(m.Player)

	eventsystem.EventManager.RegisterListener(&m, eventsystem.Stairs)
	eventsystem.EventManager.RegisterListener(&m, eventsystem.DropItem)

	// Event System
	return &m, nil
}

func (s *MainState) Done() bool {
	return false
}

func (s *MainState) Update() state.StateInterface {
	s.gui.Update(s)
	s.inventoryView.Update()
	s.tick++
	if !s.pause {
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

		if s.tick%20 == 0 {
			for _, entity := range s.levels[s.CurrentLevel].Entities {
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

	levelImage := s.levels[s.CurrentLevel].Render(s.CameraX, s.CameraY, config.GameWidth/config.SpriteWidth, config.GameHeight/config.SpriteHeight, false, true)
	op := &ebiten.DrawImageOptions{}
	//op.GeoM.Scale(1.5, 1.5)
	screen.DrawImage(levelImage, op)
	//ebitenutil.DrawRect(screen, config.GameWidth+1, 0, config.ScreenWidth-config.GameWidth-1, config.ScreenHeight, color.White)

	s.gui.Draw(screen, s)
	s.inventoryView.Draw(screen)

}

func (s *MainState) DrawPlayerMessages(screen *ebiten.Image) {
	// Player messages
	if s.Player != nil {
		x := 0
		y := 0
		for _, v := range message.MessageLog {
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

// GetMinimap
// Generates a minimap image of specified size and returns the image.
// Width and Height are in tiles not pixels.
func (g *MainState) GetMinimap(sX int, sY int, width int, height int, imageWidth int, imageHeight int) *ebiten.Image {
	worldImage := ebiten.NewImage(imageWidth, imageHeight)
	pc := g.Player.GetComponent("PositionComponent").(*component.PositionComponent)

	view := g.levels[g.CurrentLevel].GetView(sX, sY, width, height, false, false)
	for x := 0; x < len(view); x++ {
		for y := 0; y < len(view[x]); y++ {
			tX := float64(x * imageWidth / width)
			tY := float64(y * imageHeight / height)
			tile := view[x][y]

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tX, tY)
			//op.GeoM.Scale(float64(config.TileSizeW/config.SpriteSizeW), float64(config.TileSizeH/config.SpriteSizeH))

			if tile == nil {
				sX := 19 * config.SpriteWidth
				worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)
				continue
			} else {
				sX := tile.TileIndex * config.SpriteWidth
				if !tile.Seen {
					sX = 19 * config.SpriteWidth
				}
				worldImage.DrawImage(resource.Textures["map"].SubImage(image.Rect(sX, 0, sX+config.SpriteWidth, config.SpriteHeight)).(*ebiten.Image), op)

			}
		}
	}

	ebitenutil.DrawRect(worldImage, float64(pc.GetX()*imageWidth/width), float64(pc.GetY()*imageHeight/height), 5, 5, color.RGBA{0, 0, 255, 255})

	return worldImage
}

func (g *MainState) HandleEvent(data event.EventData) error {

	switch data.GetType() {
	case eventsystem.Stairs:
		stairsEvent := data.(eventsystem.StairsEventData)
		g.levels[g.CurrentLevel].DeleteEntity(g.Player)

		if stairsEvent.Up {
			if g.CurrentLevel < len(g.levels)-1 {
				g.CurrentLevel++
			}
		} else {
			if g.CurrentLevel > 0 {
				g.CurrentLevel--
			}
		}

		g.levels[g.CurrentLevel].AddEntity(g.Player)

		g.UpdateEntities()
	case eventsystem.DropItem:
		dropItemEvent := data.(eventsystem.DropItemEventData)
		//TODO Validation please
		dropItemEvent.Item.GetComponent("PositionComponent").(*component.PositionComponent).SetPosition(dropItemEvent.X, dropItemEvent.Y)
		g.levels[g.CurrentLevel].AddEntity(dropItemEvent.Item)
	}

	return nil
}
