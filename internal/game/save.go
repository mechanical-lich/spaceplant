package game

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/gamemaster"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/system"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const saveVersion = 1
const noEntity = -1

// SaveFile is the top-level JSON structure written to disk.
type SaveFile struct {
	Version      int                      `json:"version"`
	CurrentZ     int                      `json:"currentZ"`
	TickCount    int                      `json:"tickCount"`
	TurnCount    int                      `json:"turnCount"`
	PlayerID     int                      `json:"playerID"`
	Hour         int                      `json:"hour"`
	Day          int                      `json:"day"`
	Width        int                      `json:"width"`
	Height       int                      `json:"height"`
	Depth        int                      `json:"depth"`
	Tiles        []saveTile               `json:"tiles"`
	Seen         []bool                   `json:"seen"`
	NoBudding    []bool                   `json:"noBudding"`
	Overgrown    []bool                   `json:"overgrown"`
	FloorResults []generation.FloorResult `json:"floorResults"`
	Entities     []saveEntity             `json:"entities"`
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

// SaveGame serializes the current SimWorld to a JSON file at path.
func SaveGame(sw *SimWorld, path string) error {
	sw.Mu.RLock()
	defer sw.Mu.RUnlock()

	level := sw.Level

	// Assign sequential IDs to all entities (level + inventory items).
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

	for _, e := range level.Entities {
		collectEntity(e, false, false)
	}
	for _, e := range level.StaticEntities {
		collectEntity(e, true, false)
	}

	// Recursively collect inventory items (the loop grows as items are added).
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

	// Serialize each entity.
	savedEntities := make([]saveEntity, 0, len(allEntities))
	for _, e := range allEntities {
		se, err := serializeEntity(e, entityToID, isStaticMap[e], inInventoryMap[e])
		if err != nil {
			return fmt.Errorf("serialize entity %q: %w", e.Blueprint, err)
		}
		savedEntities = append(savedEntities, se)
	}

	// Serialize tiles (only type + variant; Idx/width/height are reconstructed).
	tiles := make([]saveTile, len(level.Data))
	for i, t := range level.Data {
		tiles[i] = saveTile{Type: t.Type, Variant: t.Variant}
	}

	playerID := noEntity
	if sw.Player != nil {
		if id, ok := entityToID[sw.Player]; ok {
			playerID = id
		}
	}

	sf := SaveFile{
		Version:      saveVersion,
		CurrentZ:     sw.CurrentZ,
		TickCount:    sw.TickCount,
		TurnCount:    sw.TurnCount,
		PlayerID:     playerID,
		Hour:         level.Hour,
		Day:          level.Day,
		Width:        level.Width,
		Height:       level.Height,
		Depth:        level.Depth,
		Tiles:        tiles,
		Seen:         append([]bool(nil), level.Seen...),
		NoBudding:    append([]bool(nil), level.NoBudding...),
		Overgrown:    append([]bool(nil), level.Overgrown...),
		FloorResults: sw.FloorResults,
		Entities:     savedEntities,
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal save: %w", err)
	}
	return os.WriteFile(path, data, 0644)
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
		// NOTE: component type constants (GetType() return values) are short names
		// like "BodyInventory", "Inventory", "ActiveConditions" — NOT the struct
		// names used as factory registration keys ("BodyInventoryComponent", etc.).
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

// LoadIntoSimWorld reads a save file and replaces the state of sw in-place.
// The caller must ensure no other goroutine is actively ticking sw.
func LoadIntoSimWorld(sw *SimWorld, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read save: %w", err)
	}
	var sf SaveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return fmt.Errorf("parse save: %w", err)
	}
	if sf.Version != saveVersion {
		return fmt.Errorf("incompatible save version %d (expected %d)", sf.Version, saveVersion)
	}

	// Rebuild the level (NewLevel sets up PathCostFunc and initialises tiles).
	spLevel := world.NewLevel(sf.Width, sf.Height, sf.Depth, world.NewDefaultTheme())
	for i, t := range sf.Tiles {
		spLevel.Level.Data[i].Type = t.Type
		spLevel.Level.Data[i].Variant = t.Variant
	}
	copy(spLevel.Level.Seen, sf.Seen)
	copy(spLevel.NoBudding, sf.NoBudding)
	copy(spLevel.Overgrown, sf.Overgrown)
	spLevel.Level.Hour = sf.Hour
	spLevel.Level.Day = sf.Day

	// Pass 1: create entity stubs (empty, correct Blueprint).
	idToEntity := make(map[int]*ecs.Entity, len(sf.Entities))
	for _, se := range sf.Entities {
		idToEntity[se.ID] = &ecs.Entity{Blueprint: se.Blueprint}
	}

	// Pass 2: restore plain components; inventory components deferred to pass 3.
	f := factory.GetFactory()
	for _, se := range sf.Entities {
		e := idToEntity[se.ID]
		for name, raw := range se.Components {
			switch name {
			case string(rlcomponents.BodyInventory), string(rlcomponents.Inventory):
				// Handled in pass 3.
			default:
				var data map[string]interface{}
				if err := json.Unmarshal(raw, &data); err != nil {
					return fmt.Errorf("entity %d component %s unmarshal: %w", se.ID, name, err)
				}
				// Component type keys are short (e.g. "Position"), but the factory
				// is registered under the struct name ("PositionComponent"). Try both.
				comp, err := f.CreateComponent(name, data)
				if err != nil {
					comp, err = f.CreateComponent(name+"Component", data)
					if err != nil {
						continue // Unknown component; skip gracefully.
					}
				}
				e.AddComponent(comp)
			}
		}
	}

	// Pass 3: resolve inventory entity references.
	for _, se := range sf.Entities {
		e := idToEntity[se.ID]

		if raw, ok := se.Components[string(rlcomponents.BodyInventory)]; ok {
			var sbi saveBodyInventory
			if err := json.Unmarshal(raw, &sbi); err != nil {
				return fmt.Errorf("entity %d BodyInventoryComponent: %w", se.ID, err)
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
				return fmt.Errorf("entity %d InventoryComponent: %w", se.ID, err)
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

	// Add non-inventory entities to the level (AddEntity handles isStatic via Inanimate).
	for _, se := range sf.Entities {
		if !se.InInventory {
			spLevel.AddEntity(idToEntity[se.ID])
		}
	}

	// Strip transient turn components so the turn cycle restarts cleanly.
	// If the player was saved mid-turn (MyTurn present), AdvanceEnergy would skip
	// them (it only grants MyTurn when the entity doesn't already have it), causing
	// the server to enter phaseRunningTick and deadlock waiting for commands.
	for _, e := range spLevel.Entities {
		e.RemoveComponent(rlcomponents.MyTurn)
		e.RemoveComponent(rlcomponents.TurnTaken)
	}

	// Sync skills for all level entities.
	for _, e := range spLevel.Entities {
		if e != nil {
			skill.SyncEquippedSkills(e)
		}
	}

	// Apply changes to sw under the write lock.
	sw.Mu.Lock()
	sw.Level = spLevel
	sw.CurrentZ = sf.CurrentZ
	sw.TickCount = sf.TickCount
	sw.TurnCount = sf.TurnCount
	sw.FloorResults = sf.FloorResults

	if sf.PlayerID != noEntity {
		sw.Player = idToEntity[sf.PlayerID]
	}
	sw.aiSystem.Watcher = sw.Player
	sw.advancedAISystem.Watcher = sw.Player
	sw.Mu.Unlock()

	// Recompute lights with the new level.
	sw.Mu.Lock()
	sw.UpdateEntities()
	sw.Mu.Unlock()

	return nil
}

// LoadGameNew creates a brand-new SimWorld from a save file (used at startup).
func LoadGameNew(path string) (*SimWorld, error) {
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

	if err := LoadIntoSimWorld(sw, path); err != nil {
		return nil, err
	}
	return sw, nil
}
