package game

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const numLevels = 6 // must match len(generation.FloorStack)

// CharacterData holds the player's choices from the character creator.
type CharacterData struct {
	Name         string
	PH           int
	AG           int
	MA           int
	CL           int
	LD           int
	HTCS         int
	CS           int
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
	Level            *world.Level
	Player           *ecs.Entity
	CurrentZ         int
	FloorResults     []generation.FloorResult
	systemManager    *ecs.SystemManager
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

	// Save identity fields — set when a station/player run is created or loaded.
	StationID   string
	StationName string
	PlayerRunID string

	// SelfDestructTurns counts down from the moment the self-destruct is armed.
	// 0 means inactive. When it reaches 0 after being active, the player dies.
	SelfDestructTurns    int
	selfDestructArmed    bool

	// MotherPlantPlaced is set true when the saboteur places their mother plant cutting.
	MotherPlantPlaced bool
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

	sw.FloorResults = generation.GenerateFloors(sw.Level)

	for z := 0; z < numLevels; z++ {
		sw.gm.Init(sw.Level, z)
		sw.gm.PlaceLockedProgression(sw.Level, pX, pY, z, z+1)
	}

	// Systems
	sw.systemManager.AddSystem(&rlsystems.StatusConditionSystem{
		ExtraStatuses: map[string]ecs.ComponentType{
			"Haste":  component.Haste,
			"Slowed": component.Slowed,
		},
	})
	sw.systemManager.AddSystem(&system.PlayerSystem{})
	sw.aiSystem = &system.AISystem{}
	sw.systemManager.AddSystem(sw.aiSystem)
	sw.advancedAISystem = &system.AdvancedAISystem{}
	sw.systemManager.AddSystem(sw.advancedAISystem)
	sw.systemManager.AddSystem(&system.MotherPlantSeedSystem{})
	sw.systemManager.AddSystem(&system.LightSystem{})
	sw.systemManager.AddSystem(&rlsystems.DoorSystem{AppearanceType: component.Appearance})

	item, _ := factory.Create("health", pX+2, pY)
	if item != nil {
		item.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+2, pY, 0)
		sw.Level.AddEntity(item)
	}

	crate, _ := factory.Create("crate", pX+3, pY)
	if crate != nil {
		crate.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+3, pY, 0)
		sw.Level.AddEntity(crate)
	}

	return sw, nil
}

// GetPlayer implements listeners.SimAccess.
func (sw *SimWorld) GetPlayer() *ecs.Entity { return sw.Player }

// BuildEvalContext constructs a wincondition.EvalContext from the current sim state.
func (sw *SimWorld) BuildEvalContext() wincondition.EvalContext {
	var live []*ecs.Entity
	for _, e := range sw.Level.Entities {
		if e != nil && !e.HasComponent(rlcomponents.Dead) {
			live = append(live, e)
		}
	}
	return wincondition.EvalContext{
		Player:            sw.Player,
		Entities:          live,
		SelfDestructArmed: sw.selfDestructArmed,
		MotherPlantPlaced: sw.MotherPlantPlaced,
	}
}

// RegenerateLevel builds a brand-new level in-place, resets the player, and
// clears turn/tick counters. Call this before starting a new game so the server
// keeps its existing *SimWorld pointer while all state is refreshed.
// A new StationID is generated; the caller should set StationName after calling.
func (sw *SimWorld) RegenerateLevel() error {
	pX := 50
	pY := 50
	newLevel := world.NewLevel(100, 100, numLevels, world.NewDefaultTheme())
	floorResults := generation.GenerateFloors(newLevel)

	gm := gamemaster.GameMaster{}
	for z := 0; z < numLevels; z++ {
		gm.Init(newLevel, z)
		gm.PlaceLockedProgression(newLevel, pX, pY, z, z+1)
	}

	item, _ := factory.Create("health", pX+2, pY)
	if item != nil {
		item.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+2, pY, 0)
		newLevel.AddEntity(item)
	}
	crate, _ := factory.Create("crate", pX+3, pY)
	if crate != nil {
		crate.GetComponent("Position").(*component.PositionComponent).SetPosition(pX+3, pY, 0)
		newLevel.AddEntity(crate)
	}

	sw.Mu.Lock()
	sw.Level = newLevel
	sw.FloorResults = floorResults
	sw.Player = nil
	sw.CurrentZ = 0
	sw.TickCount = 0
	sw.TurnCount = 0
	sw.StationID = generateID()
	sw.StationName = ""
	sw.PlayerRunID = ""
	sw.aiSystem.Watcher = nil
	sw.advancedAISystem.Watcher = nil
	sw.Mu.Unlock()
	return nil
}

// GetRLLevel implements listeners.SimAccess.
func (sw *SimWorld) GetRLLevel() *rlworld.Level { return sw.Level.Level }

// resolveSpawnLocation returns (x, y, z) for the player spawn point.
// If the class has StartingRoomTags, it searches all floors for the first room
// whose tag matches any of those tags and returns the room centre. Falls back
// to floor 0 stair position when no match is found.
func (sw *SimWorld) resolveSpawnLocation(classID string) (int, int, int) {
	if def := class.Get(classID); def != nil && len(def.StartingRoomTags) > 0 {
		wanted := make(map[string]bool, len(def.StartingRoomTags))
		for _, t := range def.StartingRoomTags {
			wanted[t] = true
		}
		for _, fr := range sw.FloorResults {
			for _, room := range fr.Rooms {
				if wanted[room.Tag] {
					cx := room.X + room.Width/2
					cy := room.Y + room.Height/2
					return cx, cy, fr.Z
				}
			}
		}
	}
	return sw.FloorResults[0].StairX, sw.FloorResults[0].StairY, 0
}

// SpawnPlayer creates the player entity from CharacterData and adds it to the world.
// Must be called exactly once, after character creation is complete.
func (sw *SimWorld) SpawnPlayer(data CharacterData) error {
	pX, pY, pZ := sw.resolveSpawnLocation(data.ClassID)

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
		sc.PH = data.PH
		sc.AG = data.AG
		sc.MA = data.MA
		sc.CL = data.CL
		sc.LD = data.LD
		sc.HTCS = data.HTCS
		sc.CS = data.CS
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

	// Apply class stat bonuses on top of the player's chosen stats.
	class.ApplyStatMods(player)

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

	// Give starting items from the class definition.
	if def := class.Get(data.ClassID); def != nil && len(def.StartingItems) > 0 {
		if player.HasComponent(component.BodyInventory) && player.HasComponent(component.Body) {
			inv := player.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
			bc := player.GetComponent(component.Body).(*rlcomponents.BodyComponent)
			for _, blueprintID := range def.StartingItems {
				item, err := factory.Create(blueprintID, 0, 0)
				if err != nil {
					continue
				}
				inv.AddItem(item)
			}
			inv.EquipAllBest(bc)
			skill.SyncEquippedSkills(player)
		}
	}

	// Reveal the entire starting floor for the Navigator's Stellar Cartography skill.
	if skill.HasSkill(player, "completed_minimap") {
		sw.Level.RevealFloor(pZ)
	}

	player.GetComponent("Position").(*component.PositionComponent).SetPosition(pX, pY, pZ)

	sw.Mu.Lock()
	defer sw.Mu.Unlock()
	sw.CurrentZ = pZ
	sw.Player = player
	sw.PlayerRunID = generateID()
	sw.aiSystem.Watcher = player
	sw.advancedAISystem.Watcher = player
	sw.Level.AddEntity(player)

	// Spawn initial mother plant on z=0 unless this is a saboteur run.
	if data.BackgroundID != "saboteur" && !sw.motherPlantExists() {
		sw.spawnMotherPlant(0)
	}

	sw.debugWinLocations()

	return nil
}

// debugWinLocations prints the x,y,z of win-condition entities and special rooms to stdout.
func (sw *SimWorld) debugWinLocations() {
	// Special rooms from floor results.
	roomTags := map[string]bool{
		"life_pod_bay":       true,
		"self_destruct_room": true,
	}
	for _, fr := range sw.FloorResults {
		for _, room := range fr.Rooms {
			if roomTags[room.Tag] {
				fmt.Printf("[DEBUG] room:%s at x=%d y=%d z=%d (w=%d h=%d)\n",
					room.Tag, room.X, room.Y, fr.Z, room.Width, room.Height)
			}
		}
	}

	// Key entities.
	targets := map[string]bool{
		"mother_plant":          true,
		"mobile_mother_plant":   true,
		"life_pod_console":      true,
		"self_destruct_console": true,
		"terminal":              true,
	}
	for _, e := range sw.Level.Level.GetEntities() {
		if e == nil || !targets[e.Blueprint] {
			continue
		}
		if !e.HasComponent("Position") {
			continue
		}
		pc := e.GetComponent("Position").(*component.PositionComponent)
		fmt.Printf("[DEBUG] entity:%s at x=%d y=%d z=%d\n", e.Blueprint, pc.GetX(), pc.GetY(), pc.GetZ())
	}
}

func (sw *SimWorld) motherPlantExists() bool {
	for _, e := range sw.Level.Level.GetEntities() {
		if e != nil && (e.Blueprint == "mother_plant" || e.Blueprint == "mobile_mother_plant") && !e.HasComponent("Dead") {
			return true
		}
	}
	return false
}

// spawnMotherPlant places a mother_plant entity at a random floor tile on floor z.
func (sw *SimWorld) spawnMotherPlant(z int) {
	for tries := 0; tries < 200; tries++ {
		x := rand.Intn(sw.Level.Width)
		y := rand.Intn(sw.Level.Height)
		tile := sw.Level.Level.GetTilePtr(x, y, z)
		if tile == nil || tile.IsSolid() || tile.Type == world.TypeOpen {
			continue
		}
		if sw.Level.GetEntityAt(x, y, z) != nil {
			continue
		}
		e, err := factory.Create("mobile_mother_plant", x, y)
		if err != nil {
			return
		}
		e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
		sw.Level.AddEntity(e)
		return
	}
}

// ConvertPlayerToCorpse strips the PlayerComponent from the player entity so it
// becomes a regular (dead) entity that persists on the station. The player run
// can then be marked dead and the station saved with the corpse in place.
func (sw *SimWorld) ConvertPlayerToCorpse() {
	if sw.Player == nil {
		return
	}
	sw.Player.RemoveComponent(component.Player)
	sw.Player = nil
	sw.aiSystem.Watcher = nil
	sw.advancedAISystem.Watcher = nil
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
