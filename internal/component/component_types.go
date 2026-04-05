package component

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// Re-export rlcomponents type constants for use throughout the codebase.
const (
	Position         = rlcomponents.Position
	Stats            = rlcomponents.Stats
	MyTurn           = rlcomponents.MyTurn
	Dead             = rlcomponents.Dead
	Description      = rlcomponents.Description
	Solid            = rlcomponents.Solid
	Inanimate        = rlcomponents.Inanimate
	Direction        = rlcomponents.Direction
	WanderAI         = rlcomponents.WanderAI
	HostileAI        = rlcomponents.HostileAI
	DefensiveAI      = rlcomponents.DefensiveAI
	Alerted          = rlcomponents.Alerted
	Haste            = rlcomponents.Haste
	Slowed           = rlcomponents.Slowed
	Food             = rlcomponents.Food
	Inventory        = rlcomponents.Inventory
	Item             = rlcomponents.Item
	Armor            = rlcomponents.Armor
	Weapon ecs.ComponentType = "WeaponComponent"
	NeverSleep       = rlcomponents.NeverSleep
	Nocturnal        = rlcomponents.Nocturnal
	Body             = rlcomponents.Body
	BodyInventory    = rlcomponents.BodyInventory
	Energy           = rlcomponents.Energy
	StatCondition    = rlcomponents.StatCondition
	DamageCondition  = rlcomponents.DamageCondition
	ActiveConditions = rlcomponents.ActiveConditions
)

// Spaceplant-specific component types (not in rlcomponents).
const (
	Appearance         ecs.ComponentType = "AppearanceComponent"
	Player             ecs.ComponentType = "PlayerComponent"
	Attack             ecs.ComponentType = "AttackComponent"
	Massive            ecs.ComponentType = "MassiveComponent"
	SpLight            ecs.ComponentType = "LightComponent" // spaceplant's own light (incompatible with rlcomponents.Light)
	Skill              ecs.ComponentType = "SkillComponent"
	Class              ecs.ComponentType = "ClassComponent"
	ItemSkills         ecs.ComponentType = "ItemSkillsComponent"
	Background         ecs.ComponentType = "BackgroundComponent"
	AdvancedAI         ecs.ComponentType = "AdvancedAIComponent"
	LayeredAppearance  ecs.ComponentType = "LayeredAppearanceComponent"
	WearableAppearance ecs.ComponentType = "WearableAppearanceComponent"
	HitLocation        ecs.ComponentType = "HitLocationComponent"
	Ammo               ecs.ComponentType = "AmmoComponent"
	Interaction                          = rlcomponents.Interaction
	Door                                 = rlcomponents.Door
)

// Type aliases for rlcomponents - these replace spaceplant's duplicate component structs.
type (
	InteractionComponent   = rlcomponents.InteractionComponent
	Trigger                = rlcomponents.Trigger
	DoorComponent          = rlcomponents.DoorComponent
	PositionComponent      = rlcomponents.PositionComponent
	MyTurnComponent        = rlcomponents.MyTurnComponent
	DeadComponent          = rlcomponents.DeadComponent
	DescriptionComponent   = rlcomponents.DescriptionComponent
	SolidComponent         = rlcomponents.SolidComponent
	InanimateComponent     = rlcomponents.InanimateComponent
	DirectionComponent     = rlcomponents.DirectionComponent
	WanderAIComponent      = rlcomponents.WanderAIComponent
	HostileAIComponent     = rlcomponents.HostileAIComponent
	DefensiveAIComponent   = rlcomponents.DefensiveAIComponent
	AlertedComponent       = rlcomponents.AlertedComponent
	FoodComponent          = rlcomponents.FoodComponent
	NeverSleepComponent    = rlcomponents.NeverSleepComponent
	NocturnalComponent     = rlcomponents.NocturnalComponent
	ItemComponent          = rlcomponents.ItemComponent
	InventoryComponent     = rlcomponents.InventoryComponent
	BodyInventoryComponent = rlcomponents.BodyInventoryComponent
	BodyComponent          = rlcomponents.BodyComponent
	BodyPart               = rlcomponents.BodyPart
	ItemSlot               = rlcomponents.ItemSlot
	EnergyComponent        = rlcomponents.EnergyComponent
)

// Re-export item slot constants.
const (
	HandSlot  = rlcomponents.HandSlot
	HeadSlot  = rlcomponents.HeadSlot
	TorsoSlot = rlcomponents.TorsoSlot
	LegsSlot  = rlcomponents.LegsSlot
	FeetSlot  = rlcomponents.FeetSlot
	BagSlot   = rlcomponents.BagSlot
)
