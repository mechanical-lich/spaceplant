package game

import (
	"math/rand"

	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/level"
	"github.com/mechanical-lich/spaceplant/utility"
)

const hostileMax = 20
const crewInitial = 10
const hostileInitial = 5

var hostiles = []string{"creeper", "viner", "scythe", "scrambler"}
var rareHostiles = []string{"abomination", "spitter"}
var crew = []string{"crewmember", "officer"}

type GameMaster struct {
	level *level.Level
}

// Init Initial the game master
func (gm *GameMaster) Init(l *level.Level) {
	gm.level = l

	//log.Println("Placing Crew")
	//Random food
	for i := 0; i < crewInitial; i++ {
		x := rand.Intn(gm.level.Width)
		y := rand.Intn(gm.level.Height)
		tile := gm.level.GetTileAt(x, y)
		tries := 0
		for tile.Solid || tile.Type == level.Type_Open || gm.level.GetEntityAt(x, y) != nil {
			x = rand.Intn(gm.level.Width)
			y = rand.Intn(gm.level.Height)
			tile = gm.level.GetTileAt(x, y)
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
			gm.level.AddEntity(entity)
		}
	}

	//log.Println("Placing hostiles")
	//Random hostiles
	for i := 0; i < hostileInitial; i++ {
		x := rand.Intn(gm.level.Width)
		y := rand.Intn(gm.level.Height)
		tile := gm.level.GetTileAt(x, y)
		tries := 0
		for tile.Solid || tile.Type == level.Type_Open || gm.level.GetEntityAt(x, y) != nil {
			x = rand.Intn(gm.level.Width)
			y = rand.Intn(gm.level.Height)
			tile = gm.level.GetTileAt(x, y)
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
			//log.Println("Spawn a rare hostile enemy!")
			blueprint = rareHostiles[utility.GetRandom(0, len(rareHostiles))]
		}
		food, err := factory.Create(blueprint, x, y)
		if err == nil {
			gm.level.AddEntity(food)
		}
	}
}

// Update Update the game master
func (gm *GameMaster) Update(pX, pY int) {
	hostileCount := 0

	// Random chance to spawn in a new enemy
	if utility.GetRandom(0, 5) > 3 {
		// Gather stats
		for _, e := range gm.level.Entities {
			if e.HasComponent("HostileAIComponent") {
				hostileCount++
			}
		}

		// Handle hostile count
		if hostileCount < hostileMax {
			x := rand.Intn(gm.level.Width)
			y := rand.Intn(gm.level.Height)
			tile := gm.level.GetTileAt(x, y)
			tries := 0
			dist := utility.Distance(pX, pY, x, y)

			for tile.Solid || tile.Type == level.Type_Open || gm.level.GetEntityAt(x, y) != nil || dist < 20 || dist > 50 {
				x = rand.Intn(gm.level.Width)
				y = rand.Intn(gm.level.Height)
				tile = gm.level.GetTileAt(x, y)
				dist = utility.Distance(pX, pY, x, y)

				tries++
				if tries > 1000 {
					break
				}
			}

			blueprint := hostiles[utility.GetRandom(0, len(hostiles))]

			if utility.GetRandom(0, 20) == 0 {
				//log.Println("Spawned a rare hostile enemy")
				blueprint = rareHostiles[utility.GetRandom(0, len(rareHostiles))]
			}
			e, err := factory.Create(blueprint, x, y)
			if err == nil {
				//log.Println("Spawning hostile", x, y, pY, pY)

				gm.level.AddEntity(e)
			}
		}
	}

}
