package game

import (
	"sync"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const numLevels = 4

// SimWorld holds the authoritative server-side game state.
// It is created once at startup and shared (via pointer) between
// MainSimState (server) and SPClientState (graphical client, same process).
type SimWorld struct {
	Level         *world.Level
	Player        *ecs.Entity
	CurrentZ      int
	systemManager *ecs.SystemManager
	gm            gamemaster.GameMaster
	// Mu guards Level against concurrent access between the server goroutine
	// (UpdateEntities writes) and the Ebiten render goroutine (Draw reads).
	Mu sync.RWMutex
}

// NewSimWorld constructs and populates the game world: level generation,
// systems, player, and initial entity placement.
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
	}

	// Temp stair gen
	sw.Level.SetTileTypeAt(pX, pY, 0, world.TypeStairsUp)
	sw.Level.Polish(0)
	sw.Level.SetTileTypeAt(pX, pY, 1, world.TypeStairsDown)
	sw.Level.Polish(1)

	// Systems
	sw.systemManager.AddSystem(system.InitiativeSystem{})
	sw.systemManager.AddSystem(system.StatusConditionSystem{})
	sw.systemManager.AddSystem(&system.PlayerSystem{})
	aiSystem := &system.AISystem{}
	sw.systemManager.AddSystem(aiSystem)
	sw.systemManager.AddSystem(&system.LightSystem{})
	sw.systemManager.AddSystem(&rlsystems.DoorSystem{AppearanceType: component.Appearance})

	// Player
	var err error
	sw.Player, err = factory.Create("player", pX, pY)
	if err != nil {
		return nil, err
	}
	aiSystem.Watcher = sw.Player
	sw.Player.GetComponent("Position").(*component.PositionComponent).SetPosition(pX, pY, 0)
	sw.Level.AddEntity(sw.Player)

	item, _ := factory.Create("health", pX+2, pY)
	if item != nil {
		item.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+2, pY, 0)
		sw.Level.AddEntity(item)
	}

	// Initial entity pass so systems set up state (e.g. initiative).
	sw.UpdateEntities()

	return sw, nil
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
