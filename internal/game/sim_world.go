package game

import (
	"sync"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const numLevels = 4

// CharacterData holds the player's choices from the character creator.
type CharacterData struct {
	Name         string
	Str          int
	Dex          int
	Con          int
	Int          int
	Wis          int
	ClassID      string
	ChosenSkills []string
	BackgroundID string
	BodyType     string // "mid" or "slim"
	BodyIndex    int    // 0-4 skin tone / style
	HairIndex    int    // 0-4 hair style; -1 = no hair
}

// SimWorld holds the authoritative server-side game state.
// It is created once at startup and shared (via pointer) between
// MainSimState (server) and SPClientState (graphical client, same process).
type SimWorld struct {
	Level         *world.Level
	Player        *ecs.Entity
	CurrentZ      int
	systemManager *ecs.SystemManager
	aiSystem         *system.AISystem
	advancedAISystem *system.AdvancedAISystem
	gm               gamemaster.GameMaster
	// TickCount is incremented each time the simulation advances by one tick.
	TickCount int
	// TurnCount is incremented each time the player takes a turn (i.e. spends energy).
	TurnCount int
	// Mu guards Level against concurrent access between the server goroutine
	// (UpdateEntities writes) and the Ebiten render goroutine (Draw reads).
	Mu sync.RWMutex
}

// NewSimWorld constructs and populates the game world: level generation and systems.
// The player is NOT created here; call SpawnPlayer after character creation.
func NewSimWorld() (*SimWorld, error) {
	sw := &SimWorld{
		systemManager: &ecs.SystemManager{},
		gm:            gamemaster.GameMaster{},
	}

	pX := 50
	pY := 50
	sw.Level = world.NewLevel(100, 100, numLevels, world.NewDefaultTheme())

	for z := 0; z < numLevels; z++ {
		switch utility.GetRandom(0, 3) {
		case 0:
			generation.GenerateStation(sw.Level, z, 100, 100)
		case 1:
			generation.GenerateRoundStation(sw.Level, z)
		case 2:
			generation.GenerateRectangleStation(sw.Level, z)
		}
		sw.Level.Polish(z)
		sw.gm.Init(sw.Level, z)
		sw.gm.PlaceLockedProgression(sw.Level, pX, pY, z, z+1)
	}

	// Temp stair gen
	sw.Level.SetTileTypeAt(pX, pY, 0, world.TypeStairsUp)
	sw.Level.Polish(0)
	sw.Level.SetTileTypeAt(pX, pY, 1, world.TypeStairsDown)
	sw.Level.Polish(1)

	// Systems
	sw.systemManager.AddSystem(system.StatusConditionSystem{})
	sw.systemManager.AddSystem(&system.PlayerSystem{})
	sw.aiSystem = &system.AISystem{}
	sw.systemManager.AddSystem(sw.aiSystem)
	sw.advancedAISystem = &system.AdvancedAISystem{}
	sw.systemManager.AddSystem(sw.advancedAISystem)
	sw.systemManager.AddSystem(&system.LightSystem{})
	sw.systemManager.AddSystem(&rlsystems.DoorSystem{AppearanceType: component.Appearance})

	item, _ := factory.Create("health", pX+2, pY)
	if item != nil {
		item.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+2, pY, 0)
		sw.Level.AddEntity(item)
	}

	return sw, nil
}

// GetPlayer implements listeners.SimAccess.
func (sw *SimWorld) GetPlayer() *ecs.Entity { return sw.Player }

// GetRLLevel implements listeners.SimAccess.
func (sw *SimWorld) GetRLLevel() *rlworld.Level { return sw.Level.Level }

// SpawnPlayer creates the player entity from CharacterData and adds it to the world.
// Must be called exactly once, after character creation is complete.
func (sw *SimWorld) SpawnPlayer(data CharacterData) error {
	pX, pY := 50, 50

	player, err := factory.Create("player", pX, pY)
	if err != nil {
		return err
	}

	// Override name.
	if player.HasComponent(component.Description) {
		dc := player.GetComponent(component.Description).(*component.DescriptionComponent)
		dc.Name = data.Name
	}

	// Override stats.
	if player.HasComponent(component.Stats) {
		sc := player.GetComponent(component.Stats).(*component.StatsComponent)
		sc.Str = data.Str
		sc.Dex = data.Dex
		sc.Con = data.Con
		sc.Int = data.Int
		sc.Wis = data.Wis
	}

	// Replace class — blueprint default is discarded.
	player.RemoveComponent(component.Class)
	player.AddComponent(&component.ClassComponent{
		Classes:       []string{data.ClassID},
		UpgradePoints: 1,
		ChosenSkills:  data.ChosenSkills,
	})

	// Apply chosen class skills and background skills.
	class.SyncSkills(player)

	player.AddComponent(&component.BackgroundComponent{BackgroundID: data.BackgroundID})
	background.SyncSkills(player)

	// Layered appearance.
	bodyType := data.BodyType
	if bodyType == "" {
		bodyType = "mid"
	}
	player.AddComponent(&component.LayeredAppearanceComponent{
		BodyType:  bodyType,
		BodyIndex: data.BodyIndex,
		HairIndex: data.HairIndex,
	})

	player.GetComponent("Position").(*component.PositionComponent).SetPosition(pX, pY, 0)

	sw.Mu.Lock()
	defer sw.Mu.Unlock()
	sw.Player = player
	sw.aiSystem.Watcher = player
	sw.advancedAISystem.Watcher = player
	sw.Level.AddEntity(player)
	sw.UpdateEntities()

	return nil
}

// UpdateEntities runs one full simulation pass over all level entities.
func (sw *SimWorld) UpdateEntities() {
	system.LightSystem{}.ClearLights(sw.Level, sw.CurrentZ)
	for _, entity := range sw.Level.Entities {
		if entity == nil {
			continue
		}
		sw.systemManager.UpdateSystemsForEntity(sw.Level, entity)
	}
}
