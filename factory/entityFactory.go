package factory

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/mechanical-lich/game-engine/ecs"
	"github.com/mechanical-lich/spaceplant/component"
)

var blueprints = make(map[string][]string)

// FactoryLoad Loads the blueprints for the factory to construct entities
func FactoryLoad(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	entityName := ""
	for scanner.Scan() {
		value := scanner.Text()
		if value == "" {
			entityName = ""
			continue
		}
		if entityName == "" {
			entityName = value
			continue
		} else {
			blueprints[entityName] = append(blueprints[entityName], value)
		}
	}

	return nil
}

func Create(name string, x int, y int) (*ecs.Entity, error) {
	blueprint := blueprints[name]
	if blueprint != nil {
		entity := ecs.Entity{}
		entity.Blueprint = name
		pc := &component.PositionComponent{}
		pc.SetPosition(x, y)
		entity.AddComponent(pc)

		entity.AddComponent(&component.DirectionComponent{Direction: 0})
		for _, value := range blueprint {
			c := strings.Split(value, ":")
			params := strings.Split(c[1], ",")
			switch c[0] {
			case "AppearanceComponent":
				sx, _ := strconv.Atoi(params[0])
				sy, _ := strconv.Atoi(params[1])
				r, _ := strconv.Atoi(params[2])
				g, _ := strconv.Atoi(params[3])
				b, _ := strconv.Atoi(params[4])
				entity.AddComponent(&component.AppearanceComponent{SpriteX: int(sx), SpriteY: int(sy), R: uint8(r), G: uint8(g), B: uint8(b)})

			case "InitiativeComponent":
				dv, _ := strconv.Atoi(params[0])
				ticks, _ := strconv.Atoi(params[1])
				entity.AddComponent(&component.InitiativeComponent{DefaultValue: dv, Ticks: ticks})
			case "SolidComponent":
				entity.AddComponent(&component.SolidComponent{})
			case "PlayerComponent":
				entity.AddComponent(&component.PlayerComponent{})
			case "InanimateComponent":
				entity.AddComponent(&component.InanimateComponent{})
			case "MassiveComponent":
				entity.AddComponent(&component.MassiveComponent{})
			case "NocturnalComponent":
				entity.AddComponent(&component.NocturnalComponent{})
			case "NeverSleepComponent":
				entity.AddComponent(&component.NeverSleepComponent{})
			case "InventoryComponent":
				inv := &component.InventoryComponent{}
				for _, item := range params {
					inv.AddItem(item)
				}
				entity.AddComponent(inv)
			case "LightComponent":
				radius := 5
				brightness := 5
				r := 255
				g := 0
				b := 0
				if len(params) == 5 {
					radius, _ = strconv.Atoi(params[0])
					brightness, _ = strconv.Atoi(params[1])
					r, _ = strconv.Atoi(params[2])
					g, _ = strconv.Atoi(params[3])
					b, _ = strconv.Atoi(params[4])

				}

				entity.AddComponent(&component.LightComponent{Brightness: brightness, Radius: radius, R: r, G: g, B: b})
			case "InteractComponent":
				interact := &component.InteractComponent{}
				interact.Message = append(interact.Message, params...)
				entity.AddComponent(interact)
			case "WanderAIComponent":
				entity.AddComponent(&component.WanderAIComponent{})
			case "HostileAIComponent":
				r := 5
				if len(params) == 1 {
					r, _ = strconv.Atoi(params[0])
				}

				entity.AddComponent(&component.HostileAIComponent{SightRange: r})
			case "FoodComponent":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.FoodComponent{Amount: amount})
			case "HealthComponent":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.HealthComponent{Health: amount})
			case "DamageComponent":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.DamageComponent{Amount: amount})
			case "PoisonousComponent":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.PoisonousComponent{Duration: amount})
			case "DefensiveAIComponent":
				entity.AddComponent(&component.DefensiveAIComponent{})
			case "DescriptionComponent":
				name := params[0]
				faction := "none"
				if len(params) == 2 {
					faction = params[1]
				}

				entity.AddComponent(&component.DescriptionComponent{Name: name, Faction: faction})
			}
		}
		return &entity, nil
	}
	return nil, errors.New("no blueprint found")
}
