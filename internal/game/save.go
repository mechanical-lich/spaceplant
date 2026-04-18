package game

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const saveVersion = 2
const noEntity = -1

// StationMeta is a lightweight summary for listing stations without reading full data.
type StationMeta struct {
	StationID string    `json:"stationID"`
	Name      string    `json:"name"`
	Created   time.Time `json:"created"`
}

// PlayerRunMeta is a lightweight summary for listing player runs.
type PlayerRunMeta struct {
	PlayerRunID string `json:"playerRunID"`
	StationID   string `json:"stationID"`
	Name        string `json:"name"`
	ClassName   string `json:"className"`
	Dead        bool   `json:"dead"`
	CurrentZ    int    `json:"currentZ"`
}

// StationSaveFile is the JSON structure for a persisted station (world without player).
type StationSaveFile struct {
	Version      int                      `json:"version"`
	StationID    string                   `json:"stationID"`
	Name         string                   `json:"name"`
	Width        int                      `json:"width"`
	Height       int                      `json:"height"`
	Depth        int                      `json:"depth"`
	Tiles        []saveTile               `json:"tiles"`
	NoBudding    []bool                   `json:"noBudding"`
	Overgrown    []bool                   `json:"overgrown"`
	FloorResults []generation.FloorResult `json:"floorResults"`
	Entities     []saveEntity             `json:"entities"`
	Hour         int                      `json:"hour"`
	Day          int                      `json:"day"`
}

// PlayerSaveFile is the JSON structure for a single player run on a station.
type PlayerSaveFile struct {
	Version     int          `json:"version"`
	PlayerRunID string       `json:"playerRunID"`
	StationID   string       `json:"stationID"`
	Name        string       `json:"name"`
	ClassName   string       `json:"className"`
	Dead        bool         `json:"dead"`
	CurrentZ    int          `json:"currentZ"`
	TickCount   int          `json:"tickCount"`
	TurnCount   int          `json:"turnCount"`
	Seen        []bool       `json:"seen"`
	Entities    []saveEntity `json:"entities"` // player entity + inventory items
}

type saveTile struct {
	Type    int `json:"type"`
	Variant int `json:"variant"`
}

type saveEntity struct {
	ID          int                        `json:"id"`
	Blueprint   string                     `json:"blueprint"`
	IsStatic    bool                       `json:"isStatic"`
	InInventory bool                       `json:"inInventory"`
	Components  map[string]json.RawMessage `json:"components"`
}

// saveBodyInventory replaces *ecs.Entity pointers with integer IDs.
type saveBodyInventory struct {
	Equipped          map[string]int `json:"equipped"`
	Bag               []int          `json:"bag"`
	StartingInventory []string       `json:"startingInventory"`
}

// saveInventory replaces *ecs.Entity pointers with integer IDs.
type saveInventory struct {
	LeftHand          int      `json:"leftHand"`
	RightHand         int      `json:"rightHand"`
	Head              int      `json:"head"`
	Torso             int      `json:"torso"`
	Legs              int      `json:"legs"`
	Feet              int      `json:"feet"`
	Bag               []int    `json:"bag"`
	StartingInventory []string `json:"startingInventory"`
}

// generateID generates a random 8-byte hex string for use as an ID.
func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use time-based ID
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// stationDir returns the directory for a station's save files.
func stationDir(savesDir, stationID string) string {
	return filepath.Join(savesDir, "stations", stationID)
}

// playerRunPath returns the file path for a player run save.
func playerRunPath(savesDir, playerRunID string) string {
	return filepath.Join(savesDir, "players", playerRunID+".json")
}

// SaveStation serializes the station (all non-player entities + world tiles) to disk.
// The active player entity and its inventory are excluded.
func SaveStation(sw *SimWorld, savesDir string) error {
	sw.Mu.RLock()
	defer sw.Mu.RUnlock()

	if sw.StationID == "" {
		return fmt.Errorf("station has no ID — cannot save")
	}

	level := sw.Level

	entityToID := make(map[*ecs.Entity]int)
	var allEntities []*ecs.Entity
	isStaticMap := make(map[*ecs.Entity]bool)
	inInventoryMap := make(map[*ecs.Entity]bool)

	collectEntity := func(e *ecs.Entity, static, inventory bool) {
		if e == nil {
			return
		}
		if _, exists := entityToID[e]; !exists {
			entityToID[e] = len(allEntities)
			allEntities = append(allEntities, e)
			isStaticMap[e] = static
			inInventoryMap[e] = inventory
		}
	}

	// Collect station entities, excluding the active player.
	for _, e := range level.Entities {
		if e == sw.Player {
			continue
		}
		collectEntity(e, false, false)
	}
	for _, e := range level.StaticEntities {
		collectEntity(e, true, false)
	}

	// Recursively collect inventory items from station entities only.
	for i := 0; i < len(allEntities); i++ {
		e := allEntities[i]
		if e.HasComponent(component.BodyInventory) {
			inv := e.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
			for _, item := range inv.Bag {
				collectEntity(item, false, true)
			}
			for _, item := range inv.Equipped {
				collectEntity(item, false, true)
			}
		}
		if e.HasComponent(component.Inventory) {
			inv := e.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
			for _, item := range []*ecs.Entity{inv.LeftHand, inv.RightHand, inv.Head, inv.Torso, inv.Legs, inv.Feet} {
				collectEntity(item, false, true)
			}
			for _, item := range inv.Bag {
				collectEntity(item, false, true)
			}
		}
	}

	savedEntities := make([]saveEntity, 0, len(allEntities))
	for _, e := range allEntities {
		se, err := serializeEntity(e, entityToID, isStaticMap[e], inInventoryMap[e])
		if err != nil {
			return fmt.Errorf("serialize entity %q: %w", e.Blueprint, err)
		}
		savedEntities = append(savedEntities, se)
	}

	tiles := make([]saveTile, len(level.Data))
	for i, t := range level.Data {
		tiles[i] = saveTile{Type: t.Type, Variant: t.Variant}
	}

	sf := StationSaveFile{
		Version:      saveVersion,
		StationID:    sw.StationID,
		Name:         sw.StationName,
		Width:        level.Width,
		Height:       level.Height,
		Depth:        level.Depth,
		Tiles:        tiles,
		NoBudding:    append([]bool(nil), level.NoBudding...),
		Overgrown:    append([]bool(nil), level.Overgrown...),
		FloorResults: sw.FloorResults,
		Entities:     savedEntities,
		Hour:         level.Hour,
		Day:          level.Day,
	}

	dir := stationDir(savesDir, sw.StationID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create station dir: %w", err)
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal station: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "station.json"), data, 0644); err != nil {
		return err
	}

	// Write/refresh meta.json (lightweight listing data).
	meta := StationMeta{
		StationID: sw.StationID,
		Name:      sw.StationName,
		Created:   time.Now(),
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal station meta: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), metaData, 0644)
}

// SavePlayerRun serializes the active player entity and run metadata to disk.
func SavePlayerRun(sw *SimWorld, savesDir string) error {
	sw.Mu.RLock()
	defer sw.Mu.RUnlock()

	if sw.Player == nil {
		return nil // Nothing to save if no active player.
	}
	if sw.PlayerRunID == "" {
		return fmt.Errorf("player run has no ID — cannot save")
	}

	entityToID := make(map[*ecs.Entity]int)
	var allEntities []*ecs.Entity
	inInventoryMap := make(map[*ecs.Entity]bool)

	collectEntity := func(e *ecs.Entity, inventory bool) {
		if e == nil {
			return
		}
		if _, exists := entityToID[e]; !exists {
			entityToID[e] = len(allEntities)
			allEntities = append(allEntities, e)
			inInventoryMap[e] = inventory
		}
	}

	// Start with the player entity.
	collectEntity(sw.Player, false)

	// Recursively collect player's inventory.
	for i := 0; i < len(allEntities); i++ {
		e := allEntities[i]
		if e.HasComponent(component.BodyInventory) {
			inv := e.GetComponent(component.BodyInventory).(*rlcomponents.BodyInventoryComponent)
			for _, item := range inv.Bag {
				collectEntity(item, true)
			}
			for _, item := range inv.Equipped {
				collectEntity(item, true)
			}
		}
		if e.HasComponent(component.Inventory) {
			inv := e.GetComponent(component.Inventory).(*rlcomponents.InventoryComponent)
			for _, item := range []*ecs.Entity{inv.LeftHand, inv.RightHand, inv.Head, inv.Torso, inv.Legs, inv.Feet} {
				collectEntity(item, true)
			}
			for _, item := range inv.Bag {
				collectEntity(item, true)
			}
		}
	}

	savedEntities := make([]saveEntity, 0, len(allEntities))
	for _, e := range allEntities {
		se, err := serializeEntity(e, entityToID, false, inInventoryMap[e])
		if err != nil {
			return fmt.Errorf("serialize player entity %q: %w", e.Blueprint, err)
		}
		savedEntities = append(savedEntities, se)
	}

	// Determine player name and class for display.
	name := sw.PlayerRunID
	if sw.Player.HasComponent(component.Description) {
		dc := sw.Player.GetComponent(component.Description).(*component.DescriptionComponent)
		name = dc.Name
	}

	className := ""
	if sw.Player.HasComponent(component.Class) {
		cc := sw.Player.GetComponent(component.Class).(*component.ClassComponent)
		if len(cc.Classes) > 0 {
			if def := class.Get(cc.Classes[0]); def != nil {
				className = def.Name
			} else {
				className = cc.Classes[0]
			}
		}
	}

	pf := PlayerSaveFile{
		Version:     saveVersion,
		PlayerRunID: sw.PlayerRunID,
		StationID:   sw.StationID,
		Name:        name,
		ClassName:   className,
		Dead:        false,
		CurrentZ:    sw.CurrentZ,
		TickCount:   sw.TickCount,
		TurnCount:   sw.TurnCount,
		Seen:        append([]bool(nil), sw.Level.Seen...),
		Entities:    savedEntities,
	}

	dir := filepath.Join(savesDir, "players")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create players dir: %w", err)
	}

	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal player run: %w", err)
	}
	return os.WriteFile(playerRunPath(savesDir, sw.PlayerRunID), data, 0644)
}

// SaveAll saves both the station and the active player run.
func SaveAll(sw *SimWorld, savesDir string) error {
	if err := SaveStation(sw, savesDir); err != nil {
		return fmt.Errorf("save station: %w", err)
	}
	if err := SavePlayerRun(sw, savesDir); err != nil {
		return fmt.Errorf("save player run: %w", err)
	}
	return nil
}

// MarkPlayerRunDead marks the player run file as dead without requiring a loaded SimWorld.
// Used after ConvertPlayerToCorpse to persist the dead state.
func MarkPlayerRunDead(savesDir, playerRunID string) error {
	path := playerRunPath(savesDir, playerRunID)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pf PlayerSaveFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return err
	}
	pf.Dead = true
	out, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// ListStations scans the saves directory and returns lightweight metadata for all stations.
func ListStations(savesDir string) ([]StationMeta, error) {
	stationsDir := filepath.Join(savesDir, "stations")
	entries, err := os.ReadDir(stationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var stations []StationMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(stationsDir, entry.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta StationMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		stations = append(stations, meta)
	}
	return stations, nil
}

// ListPlayerRuns returns metadata for all player runs on the given station.
func ListPlayerRuns(savesDir, stationID string) ([]PlayerRunMeta, error) {
	playersDir := filepath.Join(savesDir, "players")
	entries, err := os.ReadDir(playersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var runs []PlayerRunMeta
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(playersDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var pf PlayerSaveFile
		if err := json.Unmarshal(data, &pf); err != nil {
			continue
		}
		if pf.StationID != stationID {
			continue
		}
		runs = append(runs, PlayerRunMeta{
			PlayerRunID: pf.PlayerRunID,
			StationID:   pf.StationID,
			Name:        pf.Name,
			ClassName:   pf.ClassName,
			Dead:        pf.Dead,
			CurrentZ:    pf.CurrentZ,
		})
	}
	return runs, nil
}

// LoadStationIntoSimWorld loads a station from disk into sw (no player).
func LoadStationIntoSimWorld(sw *SimWorld, stationID, savesDir string) error {
	path := filepath.Join(stationDir(savesDir, stationID), "station.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read station: %w", err)
	}
	var sf StationSaveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return fmt.Errorf("parse station: %w", err)
	}
	if sf.Version != saveVersion {
		return fmt.Errorf("incompatible station version %d (expected %d)", sf.Version, saveVersion)
	}

	spLevel := world.NewLevel(sf.Width, sf.Height, sf.Depth, world.NewDefaultTheme())
	for i, t := range sf.Tiles {
		spLevel.Level.Data[i].Type = t.Type
		spLevel.Level.Data[i].Variant = t.Variant
	}
	copy(spLevel.NoBudding, sf.NoBudding)
	copy(spLevel.Overgrown, sf.Overgrown)
	spLevel.Level.Hour = sf.Hour
	spLevel.Level.Day = sf.Day

	idToEntity, err := deserializeEntities(sf.Entities)
	if err != nil {
		return err
	}

	addEntitiesToLevel(spLevel, sf.Entities, idToEntity)

	sw.Mu.Lock()
	sw.Level = spLevel
	sw.FloorResults = sf.FloorResults
	sw.StationID = sf.StationID
	sw.StationName = sf.Name
	sw.Player = nil
	sw.CurrentZ = 0
	sw.TickCount = 0
	sw.TurnCount = 0
	sw.aiSystem.Watcher = nil
	sw.advancedAISystem.Watcher = nil
	sw.Mu.Unlock()

	sw.Mu.Lock()
	sw.UpdateEntities()
	sw.Mu.Unlock()
	return nil
}

// LoadPlayerRunIntoSimWorld loads a player run onto the already-loaded station in sw.
// Call LoadStationIntoSimWorld first.
func LoadPlayerRunIntoSimWorld(sw *SimWorld, playerRunID, savesDir string) error {
	data, err := os.ReadFile(playerRunPath(savesDir, playerRunID))
	if err != nil {
		return fmt.Errorf("read player run: %w", err)
	}
	var pf PlayerSaveFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return fmt.Errorf("parse player run: %w", err)
	}
	if pf.Version != saveVersion {
		return fmt.Errorf("incompatible player version %d (expected %d)", pf.Version, saveVersion)
	}

	idToEntity, err := deserializeEntities(pf.Entities)
	if err != nil {
		return err
	}

	// Add the player entity (ID 0) and its inventory to the level.
	// Player entity has InInventory=false.
	sw.Mu.Lock()
	defer sw.Mu.Unlock()

	copy(sw.Level.Seen, pf.Seen)
	sw.CurrentZ = pf.CurrentZ
	sw.TickCount = pf.TickCount
	sw.TurnCount = pf.TurnCount
	sw.PlayerRunID = pf.PlayerRunID

	for _, se := range pf.Entities {
		if !se.InInventory {
			sw.Level.AddEntity(idToEntity[se.ID])
		}
	}

	// Strip transient turn components.
	for _, e := range sw.Level.Entities {
		e.RemoveComponent(rlcomponents.MyTurn)
		e.RemoveComponent(rlcomponents.TurnTaken)
	}

	// Identify the player entity (first non-inventory entity).
	for _, se := range pf.Entities {
		if !se.InInventory {
			e := idToEntity[se.ID]
			if e.HasComponent(component.Player) {
				sw.Player = e
				break
			}
		}
	}

	sw.aiSystem.Watcher = sw.Player
	sw.advancedAISystem.Watcher = sw.Player

	// Sync skills for the player.
	if sw.Player != nil {
		skill.SyncEquippedSkills(sw.Player)
	}

	return nil
}

// LoadFullGame loads both station and player run into sw.
func LoadFullGame(sw *SimWorld, stationID, playerRunID, savesDir string) error {
	if err := LoadStationIntoSimWorld(sw, stationID, savesDir); err != nil {
		return err
	}
	if err := LoadPlayerRunIntoSimWorld(sw, playerRunID, savesDir); err != nil {
		return err
	}
	return nil
}

// deserializeEntities performs the 3-pass deserialization and returns the id->entity map.
func deserializeEntities(savedEntities []saveEntity) (map[int]*ecs.Entity, error) {
	// Pass 1: create entity stubs.
	idToEntity := make(map[int]*ecs.Entity, len(savedEntities))
	for _, se := range savedEntities {
		idToEntity[se.ID] = &ecs.Entity{Blueprint: se.Blueprint}
	}

	// Pass 2: restore plain components.
	f := factory.GetFactory()
	for _, se := range savedEntities {
		e := idToEntity[se.ID]
		for name, raw := range se.Components {
			switch name {
			case string(rlcomponents.BodyInventory), string(rlcomponents.Inventory):
				// Handled in pass 3.
			default:
				var data map[string]interface{}
				if err := json.Unmarshal(raw, &data); err != nil {
					return nil, fmt.Errorf("entity %d component %s unmarshal: %w", se.ID, name, err)
				}
				comp, err := f.CreateComponent(name, data)
				if err != nil {
					comp, err = f.CreateComponent(name+"Component", data)
					if err != nil {
						continue
					}
				}
				e.AddComponent(comp)
			}
		}
	}

	// Pass 3: resolve inventory entity references.
	for _, se := range savedEntities {
		e := idToEntity[se.ID]

		if raw, ok := se.Components[string(rlcomponents.BodyInventory)]; ok {
			var sbi saveBodyInventory
			if err := json.Unmarshal(raw, &sbi); err != nil {
				return nil, fmt.Errorf("entity %d BodyInventoryComponent: %w", se.ID, err)
			}
			inv := &rlcomponents.BodyInventoryComponent{
				StartingInventory: sbi.StartingInventory,
				Equipped:          make(map[string]*ecs.Entity, len(sbi.Equipped)),
				Bag:               make([]*ecs.Entity, 0, len(sbi.Bag)),
			}
			for part, id := range sbi.Equipped {
				if id != noEntity {
					inv.Equipped[part] = idToEntity[id]
				}
			}
			for _, id := range sbi.Bag {
				if id != noEntity {
					inv.Bag = append(inv.Bag, idToEntity[id])
				}
			}
			e.AddComponent(inv)
		}

		if raw, ok := se.Components[string(rlcomponents.Inventory)]; ok {
			var si saveInventory
			if err := json.Unmarshal(raw, &si); err != nil {
				return nil, fmt.Errorf("entity %d InventoryComponent: %w", se.ID, err)
			}
			inv := &rlcomponents.InventoryComponent{
				StartingInventory: si.StartingInventory,
				Bag:               make([]*ecs.Entity, 0, len(si.Bag)),
			}
			if si.LeftHand != noEntity {
				inv.LeftHand = idToEntity[si.LeftHand]
			}
			if si.RightHand != noEntity {
				inv.RightHand = idToEntity[si.RightHand]
			}
			if si.Head != noEntity {
				inv.Head = idToEntity[si.Head]
			}
			if si.Torso != noEntity {
				inv.Torso = idToEntity[si.Torso]
			}
			if si.Legs != noEntity {
				inv.Legs = idToEntity[si.Legs]
			}
			if si.Feet != noEntity {
				inv.Feet = idToEntity[si.Feet]
			}
			for _, id := range si.Bag {
				if id != noEntity {
					inv.Bag = append(inv.Bag, idToEntity[id])
				}
			}
			e.AddComponent(inv)
		}
	}

	return idToEntity, nil
}

// addEntitiesToLevel adds non-inventory entities from a deserialized list to a level.
func addEntitiesToLevel(level *world.Level, savedEntities []saveEntity, idToEntity map[int]*ecs.Entity) {
	for _, se := range savedEntities {
		if !se.InInventory {
			level.AddEntity(idToEntity[se.ID])
		}
	}

	// Strip transient turn components.
	for _, e := range level.Entities {
		e.RemoveComponent(rlcomponents.MyTurn)
		e.RemoveComponent(rlcomponents.TurnTaken)
	}

	// Sync skills.
	for _, e := range level.Entities {
		if e != nil {
			skill.SyncEquippedSkills(e)
		}
	}
}

func serializeEntity(e *ecs.Entity, entityToID map[*ecs.Entity]int, static, inventory bool) (saveEntity, error) {
	se := saveEntity{
		ID:          entityToID[e],
		Blueprint:   e.Blueprint,
		IsStatic:    static,
		InInventory: inventory,
		Components:  make(map[string]json.RawMessage),
	}

	for compType, comp := range e.Components {
		name := string(compType)
		switch name {
		case string(rlcomponents.BodyInventory):
			inv := comp.(*rlcomponents.BodyInventoryComponent)
			sbi := saveBodyInventory{
				StartingInventory: inv.StartingInventory,
				Equipped:          make(map[string]int, len(inv.Equipped)),
				Bag:               make([]int, 0, len(inv.Bag)),
			}
			for part, item := range inv.Equipped {
				sbi.Equipped[part] = entityRefID(item, entityToID)
			}
			for _, item := range inv.Bag {
				if item != nil {
					sbi.Bag = append(sbi.Bag, entityRefID(item, entityToID))
				}
			}
			raw, err := json.Marshal(sbi)
			if err != nil {
				return se, err
			}
			se.Components[name] = raw

		case string(rlcomponents.Inventory):
			inv := comp.(*rlcomponents.InventoryComponent)
			si := saveInventory{
				StartingInventory: inv.StartingInventory,
				LeftHand:          entityRefID(inv.LeftHand, entityToID),
				RightHand:         entityRefID(inv.RightHand, entityToID),
				Head:              entityRefID(inv.Head, entityToID),
				Torso:             entityRefID(inv.Torso, entityToID),
				Legs:              entityRefID(inv.Legs, entityToID),
				Feet:              entityRefID(inv.Feet, entityToID),
				Bag:               make([]int, 0, len(inv.Bag)),
			}
			for _, item := range inv.Bag {
				if item != nil {
					si.Bag = append(si.Bag, entityRefID(item, entityToID))
				}
			}
			raw, err := json.Marshal(si)
			if err != nil {
				return se, err
			}
			se.Components[name] = raw

		case string(component.Player):
			// Save a clean PlayerComponent — pending transient actions are discarded.
			raw, err := json.Marshal(&component.PlayerComponent{})
			if err != nil {
				return se, err
			}
			se.Components[name] = raw

		case string(rlcomponents.ActiveConditions):
			// Transient status effects are not persisted.

		default:
			raw, err := json.Marshal(comp)
			if err != nil {
				return se, fmt.Errorf("component %s: %w", name, err)
			}
			se.Components[name] = raw
		}
	}

	return se, nil
}

func entityRefID(e *ecs.Entity, entityToID map[*ecs.Entity]int) int {
	if e == nil {
		return noEntity
	}
	if id, ok := entityToID[e]; ok {
		return id
	}
	return noEntity
}

// LoadGameNew creates a brand-new SimWorld from a full save (station + player run).
// Used when both IDs are known at startup.
func LoadGameNew(stationID, playerRunID, savesDir string) (*SimWorld, error) {
	sw := &SimWorld{
		systemManager: &ecs.SystemManager{},
		gm:            gamemaster.GameMaster{},
	}

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
	sw.systemManager.AddSystem(&system.LightSystem{})
	sw.systemManager.AddSystem(&rlsystems.DoorSystem{AppearanceType: component.Appearance})

	if err := LoadFullGame(sw, stationID, playerRunID, savesDir); err != nil {
		return nil, err
	}
	return sw, nil
}
