package gamemaster

import (
	"fmt"
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/scenario"
	"github.com/mechanical-lich/spaceplant/internal/stationconfig"
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// pickTile selects a random unoccupied passable tile from candidates, falling back
// to a full random search if candidates is empty.
func pickTile(l *world.Level, z int, candidates [][2]int) (x, y int, ok bool) {
	// Shuffle candidates and return the first unoccupied one.
	if len(candidates) > 0 {
		perm := rand.Perm(len(candidates))
		for _, i := range perm {
			cx, cy := candidates[i][0], candidates[i][1]
			if l.GetEntityAt(cx, cy, z) == nil {
				return cx, cy, true
			}
		}
		return 0, 0, false
	}
	// No candidates list — random search fallback.
	for tries := 0; tries < 200; tries++ {
		rx := rand.Intn(l.Width)
		ry := rand.Intn(l.Height)
		t := l.Level.GetTilePtr(rx, ry, z)
		if t == nil || t.IsSolid() || t.Type == world.TypeOpen {
			continue
		}
		if l.GetEntityAt(rx, ry, z) != nil {
			continue
		}
		return rx, ry, true
	}
	return 0, 0, false
}

// crewInitial is now read from stationconfig at runtime.

var crew = []string{"crewmember", "officer"}

type GameMaster struct {
}

// Init initialises the game master for floor z, spawning crew and scenario hostiles.
func (gm *GameMaster) Init(l *world.Level, z int, fr generation.FloorResult) {
	themeName := ""
	if fr.Theme != nil {
		themeName = fr.Theme.Name
	}

	// Crew — random tile anywhere on the floor.
	for i := 0; i < stationconfig.Get().CrewCapacity; i++ {
		x, y, ok := pickTile(l, z, nil)
		if !ok {
			continue
		}
		blueprint := crew[utility.GetRandom(0, len(crew))]
		entity, err := factory.Create(blueprint, x, y)
		if err == nil {
			entity.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
			l.AddEntity(entity)
		}
	}

	s := scenario.Active()

	// Filter hostile pool to blueprints whose floor rule matches this floor.
	validHostiles := scenario.FilterByFloor(s.Hostiles, s.SpawnRules, z, themeName)
	validRare := scenario.FilterByFloor(s.RareHostiles, s.SpawnRules, z, themeName)

	// Skip hostile spawning on this floor if no blueprints are valid here.
	if len(validHostiles) == 0 && len(validRare) == 0 {
		return
	}
	if len(validHostiles) == 0 {
		validHostiles = validRare
	}
	if len(validRare) == 0 {
		validRare = validHostiles
	}

	for i := 0; i < s.HostileInitial; i++ {
		blueprint := validHostiles[utility.GetRandom(0, len(validHostiles))]
		if utility.GetRandom(0, 30) == 0 {
			blueprint = validRare[utility.GetRandom(0, len(validRare))]
		}
		rule := s.SpawnRules[blueprint]
		candidates := scenario.SpawnTiles(l, z, fr, rule)
		x, y, ok := pickTile(l, z, candidates)
		if !ok {
			continue
		}
		e, err := factory.Create(blueprint, x, y)
		if err == nil {
			e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
			l.AddEntity(e)
		}
	}
}

// PlaceLockedProgression locks up to numLocks doors on layer z and places a
// matching key in the reachable area before each door is locked, preventing
// soft-locks by construction.
func (gm *GameMaster) PlaceLockedProgression(l *world.Level, spawnX, spawnY, z, numLocks int) {
	// Collect all door entities on this Z layer.
	var doors []*ecs.Entity
	for _, e := range l.Level.GetEntities() {
		if e == nil || !e.HasComponent(rlcomponents.Door) || !e.HasComponent(rlcomponents.Position) {
			continue
		}
		pc := e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		if pc.GetZ() == z {
			doors = append(doors, e)
		}
	}

	rand.Shuffle(len(doors), func(i, j int) { doors[i], doors[j] = doors[j], doors[i] })
	if numLocks < len(doors) {
		doors = doors[:numLocks]
	}

	lockedDoors := map[*ecs.Entity]bool{}

	for i, door := range doors {
		blocked := map[*ecs.Entity]bool{door: true}
		for d := range lockedDoors {
			blocked[d] = true
		}

		reachable := generation.FloodFillReachable(l, spawnX, spawnY, z, blocked)

		// Filter to tiles with a passable floor type and no entity.
		var valid [][2]int
		for _, coord := range reachable {
			x, y := coord[0], coord[1]
			tile := l.Level.GetTilePtr(x, y, z)
			if tile == nil {
				continue
			}
			if tile.Type != world.TypeFloor && tile.Type != world.TypeMaintenanceTunnelFloor {
				continue
			}
			if l.Level.GetEntityAt(x, y, z) != nil {
				continue
			}
			valid = append(valid, coord)
		}

		if len(valid) == 0 {
			continue
		}

		pick := valid[rand.Intn(len(valid))]
		keyID := fmt.Sprintf("key_%d_%d", z, i)
		createKeyEntity(l, pick[0], pick[1], z, keyID)

		// Lock the door.
		dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
		dc.Locked = true
		dc.KeyId = keyID

		// Red tint so the player can see it is locked.
		if door.HasComponent(component.Appearance) {
			ac := door.GetComponent(component.Appearance).(*component.AppearanceComponent)
			ac.R = 220
			ac.G = 80
			ac.B = 80
		}

		lockedDoors[door] = true
	}
}

// createKeyEntity places a key entity that unlocks the door with the given keyID.
func createKeyEntity(l *world.Level, x, y, z int, keyID string) {
	e := &ecs.Entity{}
	pc := &rlcomponents.PositionComponent{}
	pc.SetPosition(x, y, z)
	e.AddComponent(pc)
	e.AddComponent(&rlcomponents.KeyComponent{KeyID: keyID})
	e.AddComponent(&rlcomponents.ItemComponent{Effect: "key", Value: 10, Slot: rlcomponents.BagSlot})
	e.AddComponent(&rlcomponents.DescriptionComponent{
		Name:                "Keycard",
		Faction:             "item",
		PassOverDescription: []string{"A keycard lies here."},
	})
	e.AddComponent(&component.AppearanceComponent{
		SpriteX:  0,
		SpriteY:  0,
		Resource: "keys",
		R:        200,
		G:        200,
		B:        125,
	})
	l.Level.AddEntity(e)
}

// Update runs per-tick game master logic, potentially spawning a new hostile.
func (gm *GameMaster) Update(l *world.Level, z, pX, pY int, fr generation.FloorResult) {
	s := scenario.Active()

	spawnRoll := float64(utility.GetRandom(0, 100)) / 100.0
	if spawnRoll >= s.SpawnChance {
		return
	}

	hostileCount := 0
	for _, e := range l.Entities {
		if e.HasComponent("HostileAI") {
			hostileCount++
		}
	}
	if hostileCount >= s.HostileMax {
		return
	}

	themeName := ""
	if fr.Theme != nil {
		themeName = fr.Theme.Name
	}
	validHostiles := scenario.FilterByFloor(s.Hostiles, s.SpawnRules, z, themeName)
	validRare := scenario.FilterByFloor(s.RareHostiles, s.SpawnRules, z, themeName)
	if len(validHostiles) == 0 && len(validRare) == 0 {
		return
	}
	if len(validHostiles) == 0 {
		validHostiles = validRare
	}
	if len(validRare) == 0 {
		validRare = validHostiles
	}

	blueprint := validHostiles[utility.GetRandom(0, len(validHostiles))]
	if utility.GetRandom(0, 20) == 0 {
		blueprint = validRare[utility.GetRandom(0, len(validRare))]
	}

	// Restrict to tiles at a sane distance from the player.
	rule := s.SpawnRules[blueprint]
	candidates := scenario.SpawnTiles(l, z, fr, rule)
	// Filter by player distance.
	filtered := candidates[:0]
	for _, c := range candidates {
		d := utility.Distance(pX, pY, c[0], c[1])
		if d >= 20 && d <= 50 {
			filtered = append(filtered, c)
		}
	}

	x, y, ok := pickTile(l, z, filtered)
	if !ok {
		return
	}
	e, err := factory.Create(blueprint, x, y)
	if err == nil {
		e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
		l.AddEntity(e)
	}
}
