package factory

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/mechanical-lich/mlge/ecs"
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
		pc.SetPosition(x, y, 0)
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
				frameCount, _ := strconv.Atoi(params[5])

				entity.AddComponent(&component.AppearanceComponent{SpriteX: int(sx), SpriteY: int(sy), R: uint8(r), G: uint8(g), B: uint8(b), FrameCount: frameCount})

			case "Initiative":
				dv, _ := strconv.Atoi(params[0])
				ticks, _ := strconv.Atoi(params[1])
				entity.AddComponent(&component.InitiativeComponent{DefaultValue: dv, Ticks: ticks})
			case "Item":
				effect := params[0]
				value, _ := strconv.Atoi(params[1])
				slot := params[2]
				entity.AddComponent(&component.ItemComponent{Effect: effect, Value: value, Slot: component.ItemSlot(slot)})
			case "Armor":
				value, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.ArmorComponent{DefenseBonus: value})
			case "Weapon":
				value, _ := strconv.Atoi(params[0])
				dice := params[1]
				entity.AddComponent(&component.WeaponComponent{AttackBonus: value, AttackDice: dice})
			case "Solid":
				entity.AddComponent(&component.SolidComponent{})
			case "PlayerComponent":
				entity.AddComponent(&component.PlayerComponent{})
			case "Inanimate":
				entity.AddComponent(&component.InanimateComponent{})
			case "MassiveComponent":
				entity.AddComponent(&component.MassiveComponent{})
			case "Nocturnal":
				entity.AddComponent(&component.NocturnalComponent{})
			case "NeverSleep":
				entity.AddComponent(&component.NeverSleepComponent{})
			case "Inventory":
				inv := &component.InventoryComponent{}
				entity.AddComponent(inv)
				if len(params) > 0 {
					importString := ""
					for _, v := range params {
						if importString != "" {
							importString += ","
						}
						importString += v
					}
					ImportInventory(importString, &entity)
				}
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
			case "WanderAI":
				entity.AddComponent(&component.WanderAIComponent{})
			case "HostileAI":
				r := 5
				if len(params) == 1 {
					r, _ = strconv.Atoi(params[0])
				}

				entity.AddComponent(&component.HostileAIComponent{SightRange: r})
			case "Food":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.FoodComponent{Amount: amount})
			case "Health":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.HealthComponent{Health: amount})
			case "Stats":
				ac, _ := strconv.Atoi(params[0])
				str, _ := strconv.Atoi(params[1])
				dex, _ := strconv.Atoi(params[2])
				intel, _ := strconv.Atoi(params[3])
				wis, _ := strconv.Atoi(params[4])
				d := params[5]
				entity.AddComponent(&component.StatsComponent{AC: ac, Str: str, Dex: dex, Int: intel, Wis: wis, BasicAttackDice: d})
			case "Poisonous":
				amount, _ := strconv.Atoi(params[0])
				entity.AddComponent(&component.PoisonousComponent{Duration: amount})
			case "DefensiveAI":
				entity.AddComponent(&component.DefensiveAIComponent{})
			case "Description":
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

func ImportInventory(importString string, target *ecs.Entity) error {
	if !target.HasComponent("Inventory") {
		return errors.New("no inventory")
	}

	inventory := target.GetComponent("Inventory").(*component.InventoryComponent)
	imports := strings.Split(importString, ",")
	for _, v := range imports {
		entity, err := Create(v, 0, 0)
		if err != nil {
			return err
		}
		inventory.AddItem(entity)
		inventory.Equip(entity)
	}

	return nil
}
