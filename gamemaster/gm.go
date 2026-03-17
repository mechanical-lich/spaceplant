package gamemaster

import (
	"math/rand"

	"github.com/mechanical-lich/spaceplant/component"
	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/utility"
	"github.com/mechanical-lich/spaceplant/world"
)

const hostileMax = 20
const crewInitial = 10
const hostileInitial = 15

var hostiles = []string{"creeper", "viner", "scythe", "scrambler"}
var rareHostiles = []string{"abomination", "spitter"}
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

	for i := 0; i < hostileInitial; i++ {
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

		blueprint := hostiles[utility.GetRandom(0, len(hostiles))]
		if utility.GetRandom(0, 100) == 0 {
			blueprint = rareHostiles[utility.GetRandom(0, len(rareHostiles))]
		}
		food, err := factory.Create(blueprint, x, y)
		if err == nil {
			food.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
			l.AddEntity(food)
		}
	}
}

// Update Update the game master
func (gm *GameMaster) Update(l *world.Level, z, pX, pY int) {
	hostileCount := 0

	// Random chance to spawn in a new enemy
	if utility.GetRandom(0, 5) > 3 {
		// Gather stats
		for _, e := range l.Entities {
			if e.HasComponent("HostileAI") {
				hostileCount++
			}
		}

		// Handle hostile count
		if hostileCount < hostileMax {
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

			blueprint := hostiles[utility.GetRandom(0, len(hostiles))]

			if utility.GetRandom(0, 20) == 0 {
				blueprint = rareHostiles[utility.GetRandom(0, len(rareHostiles))]
			}
			e, err := factory.Create(blueprint, x, y)
			if err == nil {
				e.GetComponent("Position").(*component.PositionComponent).SetPosition(x, y, z)
				l.AddEntity(e)
			}
		}
	}
}
