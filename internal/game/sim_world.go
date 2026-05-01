package game

import (
	"log"
	"math/rand"
	"os"
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
	"github.com/mechanical-lich/spaceplant/internal/scenario"
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
	systemManager *ecs.SystemManager
	gm            gamemaster.GameMaster
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

	// MotherPlantPlaced is set true when the saboteur places their mother plant cutting.
	MotherPlantPlaced bool
}

// NewSimWorld constructs and populates the game world: level generation and systems.
// The player is NOT created here; call SpawnPlayer after character creation.
func init() {
	component.SpawnEntityFunc = func(blueprint string, x, y, z int, levelData any) error {
		l, ok := levelData.(*world.Level)
		if !ok {
			return nil
		}
		e, err := factory.Create(blueprint, x, y)
		if err != nil {
			return err
		}
		e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent).SetPosition(x, y, z)
		l.AddEntity(e)
		return nil
	}
}

func NewSimWorld() (*SimWorld, error) {
	sw := &SimWorld{
		systemManager: &ecs.SystemManager{},
		gm:            gamemaster.GameMaster{},
	}

	pX := 50
	pY := 50
	sw.Level = world.NewLevel(100, 100, numLevels, world.NewDefaultTheme())
	if b, err := os.ReadFile("data/shaders/crt.kage"); err != nil {
		log.Printf("crt shader not loaded: %v", err)
	} else {
		sw.Level.ShaderSrc = b
	}

	sw.FloorResults = generation.GenerateFloors(sw.Level)

	for z := 0; z < numLevels; z++ {
		sw.gm.Init(sw.Level, z, sw.FloorResults[z])
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
	sw.systemManager.AddSystem(&system.ScriptSystem{})
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

	if s := scenario.Active(); s != nil && len(s.SetupScripts) > 0 {
		system.RunSetupScripts(s.SetupScripts, sw.Level, sw.FloorResults)
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
		Flags:             sw.Level.Flags,
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
		gm.Init(newLevel, z, floorResults[z])
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

	newLevel.ShaderSrc = sw.Level.ShaderSrc

	if s := scenario.Active(); s != nil && len(s.SetupScripts) > 0 {
		system.RunSetupScripts(s.SetupScripts, newLevel, floorResults)
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
	sw.Mu.Unlock()
	return nil
}

// GetRLLevel implements listeners.SimAccess.
func (sw *SimWorld) GetRLLevel() *rlworld.Level { return sw.Level.Level }

// SelfDestructArmed returns true when the self-destruct flag has been set by a script.
func (sw *SimWorld) SelfDestructArmed() bool {
	v := sw.Level.Flags["self_destruct_armed"]
	return v != nil && v != false && v != 0.0
}

// SelfDestructTurns returns the remaining countdown stored in Level.Flags by the script.
func (sw *SimWorld) SelfDestructTurns() int {
	v := sw.Level.Flags["self_destruct_turns"]
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// GetLevel implements listeners.SimAccess.
func (sw *SimWorld) GetLevel() *world.Level { return sw.Level }

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
	sw.Level.AddEntity(player)

	// Spawn scenario boss(es) unless the saboteur background handles placement manually.
	if data.BackgroundID != "saboteur" {
		s := scenario.Active()
		for _, bp := range s.BossSpawns {
			rule := s.SpawnRules[bp]
			// Determine which floor to use: first floor matching the rule, default z=0.
			targetZ := 0
			for _, fr := range sw.FloorResults {
				themeName := ""
				if fr.Theme != nil {
					themeName = fr.Theme.Name
				}
				if rule.FloorMatches(fr.Z, themeName) {
					targetZ = fr.Z
					break
				}
			}
			fr := sw.FloorResults[targetZ]
			candidates := scenario.SpawnTiles(sw.Level, targetZ, fr, rule)
			sw.spawnBossFromCandidates(bp, targetZ, candidates)
		}
	}

	return nil
}

func (sw *SimWorld) motherPlantExists() bool {
	for _, e := range sw.Level.Level.GetEntities() {
		if e != nil && (e.Blueprint == "mother_plant" || e.Blueprint == "mobile_mother_plant") && !e.HasComponent("Dead") {
			return true
		}
	}
	return false
}

// spawnMotherPlant places a mobile_mother_plant at a random floor tile on floor z.
// Kept for the saboteur background path.
func (sw *SimWorld) spawnMotherPlant(z int) {
	sw.spawnBossFromCandidates("mobile_mother_plant", z, nil)
}

// spawnBossFromCandidates places blueprint at a random tile from candidates.
// If candidates is nil/empty, falls back to any passable tile on floor z.
func (sw *SimWorld) spawnBossFromCandidates(blueprint string, z int, candidates [][2]int) {
	// Build fallback list if no candidates provided.
	if len(candidates) == 0 {
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
			candidates = append(candidates, [2]int{x, y})
			break
		}
	}
	for _, c := range rand.Perm(len(candidates)) {
		x, y := candidates[c][0], candidates[c][1]
		if sw.Level.GetEntityAt(x, y, z) != nil {
			continue
		}
		e, err := factory.Create(blueprint, x, y)
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
