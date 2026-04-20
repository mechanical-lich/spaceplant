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
	"github.com/mechanical-lich/spaceplant/internal/utility"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const crewInitial = 10

var crew = []string{"crewmember", "officer"}

type GameMaster struct {
}

// Init Initial the game master
func (gm *GameMaster) Init(l *world.Level, z int) {
	for i := 0; i < crewInitial; i++ {
		x := rand.Intn(l.Width)
		y := rand.Intn(l.Height)
		tile := l.Level.GetTilePtr(x, y, z)
		tries := 0
		for tile == nil || tile.IsSolid() || tile.Type == world.TypeOpen || l.GetEntityAt(x, y, z) != nil {
			x = rand.Intn(l.Width)
			y = rand.Intn(l.Height)
			tile = l.Level.GetTilePtr(x, y, z)
			tries++
			if tries > 10 {
				break
			}
		}
		if tries > 10 {
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
	for i := 0; i < s.HostileInitial; i++ {
		x := rand.Intn(l.Width)
		y := rand.Intn(l.Height)
		tile := l.Level.GetTilePtr(x, y, z)
		tries := 0
		for tile == nil || tile.IsSolid() || tile.Type == world.TypeOpen || l.GetEntityAt(x, y, z) != nil {
			x = rand.Intn(l.Width)
			y = rand.Intn(l.Height)
			tile = l.Level.GetTilePtr(x, y, z)
			tries++
			if tries > 10 {
				break
			}
		}
		if tries > 10 {
			continue
		}

		blueprint := s.Hostiles[utility.GetRandom(0, len(s.Hostiles))]
		if utility.GetRandom(0, 30) == 0 {
			blueprint = s.RareHostiles[utility.GetRandom(0, len(s.RareHostiles))]
		}
		food, err := factory.Create(blueprint, x, y)
		if err == nil {
			food.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
			l.AddEntity(food)
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

// Update Update the game master
func (gm *GameMaster) Update(l *world.Level, z, pX, pY int) {
	s := scenario.Active()
	hostileCount := 0

	spawnRoll := float64(utility.GetRandom(0, 100)) / 100.0
	if spawnRoll < s.SpawnChance {
		for _, e := range l.Entities {
			if e.HasComponent("HostileAI") {
				hostileCount++
			}
		}

		if hostileCount < s.HostileMax {
			x := rand.Intn(l.Width)
			y := rand.Intn(l.Height)
			tile := l.Level.GetTilePtr(x, y, z)
			tries := 0
			dist := utility.Distance(pX, pY, x, y)

			for tile == nil || tile.IsSolid() || tile.Type == world.TypeOpen || l.GetEntityAt(x, y, z) != nil || dist < 20 || dist > 50 {
				x = rand.Intn(l.Width)
				y = rand.Intn(l.Height)
				tile = l.Level.GetTilePtr(x, y, z)
				dist = utility.Distance(pX, pY, x, y)

				tries++
				if tries > 1000 {
					break
				}
			}

			blueprint := s.Hostiles[utility.GetRandom(0, len(s.Hostiles))]

			if utility.GetRandom(0, 20) == 0 {
				blueprint = s.RareHostiles[utility.GetRandom(0, len(s.RareHostiles))]
			}
			e, err := factory.Create(blueprint, x, y)
			if err == nil {
				e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
				l.AddEntity(e)
			}
		}
	}
}
