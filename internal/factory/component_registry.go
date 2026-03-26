package factory

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// registerComponents registers all component constructors with the jsonFactory.
func registerComponents() {
	jsonFactory.RegisterComponent("DescriptionComponent", func() ecs.Component { return &rlcomponents.DescriptionComponent{} })
jsonFactory.RegisterComponent("AppearanceComponent", func() ecs.Component { return &component.AppearanceComponent{} })
	jsonFactory.RegisterComponent("SolidComponent", func() ecs.Component { return &rlcomponents.SolidComponent{} })
	jsonFactory.RegisterComponent("InitiativeComponent", func() ecs.Component { return &rlcomponents.InitiativeComponent{} })
	jsonFactory.RegisterComponent("InventoryComponent", func() ecs.Component { return &rlcomponents.InventoryComponent{} })
	jsonFactory.RegisterComponent("StatsComponent", func() ecs.Component { return &rlcomponents.StatsComponent{} })
	jsonFactory.RegisterComponent("InanimateComponent", func() ecs.Component { return &rlcomponents.InanimateComponent{} })
	jsonFactory.RegisterComponent("PositionComponent", func() ecs.Component { return &rlcomponents.PositionComponent{} })
	jsonFactory.RegisterComponent("DirectionComponent", func() ecs.Component { return &rlcomponents.DirectionComponent{} })
	jsonFactory.RegisterComponent("WanderAIComponent", func() ecs.Component { return &rlcomponents.WanderAIComponent{} })
	jsonFactory.RegisterComponent("HostileAIComponent", func() ecs.Component { return &rlcomponents.HostileAIComponent{} })
	jsonFactory.RegisterComponent("FoodComponent", func() ecs.Component { return &rlcomponents.FoodComponent{} })
	jsonFactory.RegisterComponent("PoisonousComponent", func() ecs.Component { return &rlcomponents.PoisonousComponent{} })
	jsonFactory.RegisterComponent("PoisonedComponent", func() ecs.Component { return &rlcomponents.PoisonedComponent{} })
	jsonFactory.RegisterComponent("DefensiveAIComponent", func() ecs.Component { return &rlcomponents.DefensiveAIComponent{} })
	jsonFactory.RegisterComponent("NeverSleepComponent", func() ecs.Component { return &rlcomponents.NeverSleepComponent{} })
	jsonFactory.RegisterComponent("NocturnalComponent", func() ecs.Component { return &rlcomponents.NocturnalComponent{} })
	jsonFactory.RegisterComponent("MassiveComponent", func() ecs.Component { return &component.MassiveComponent{} })
	jsonFactory.RegisterComponent("InteractionComponent", func() ecs.Component { return &rlcomponents.InteractionComponent{} })
	jsonFactory.RegisterComponent("DoorComponent", func() ecs.Component { return &rlcomponents.DoorComponent{} })
	jsonFactory.RegisterComponent("ItemComponent", func() ecs.Component { return &rlcomponents.ItemComponent{} })
	jsonFactory.RegisterComponent("ArmorComponent", func() ecs.Component { return &rlcomponents.ArmorComponent{} })
	jsonFactory.RegisterComponent("WeaponComponent", func() ecs.Component { return &rlcomponents.WeaponComponent{} })
	jsonFactory.RegisterComponent("MyTurnComponent", func() ecs.Component { return &rlcomponents.MyTurnComponent{} })
	jsonFactory.RegisterComponent("DeadComponent", func() ecs.Component { return &rlcomponents.DeadComponent{} })
	jsonFactory.RegisterComponent("AlertedComponent", func() ecs.Component { return &rlcomponents.AlertedComponent{} })
	jsonFactory.RegisterComponent("PlayerComponent", func() ecs.Component { return &component.PlayerComponent{} })
	jsonFactory.RegisterComponent("AttackComponent", func() ecs.Component { return &component.AttackComponent{} })
	jsonFactory.RegisterComponent("LightComponent", func() ecs.Component { return &component.LightComponent{} })
	jsonFactory.RegisterComponent("AsciiAppearanceComponent", func() ecs.Component { return &rlcomponents.AsciiAppearanceComponent{} })
	jsonFactory.RegisterComponent("KeyComponent", func() ecs.Component { return &rlcomponents.KeyComponent{} })
	jsonFactory.RegisterComponent("TurnTakenComponent", func() ecs.Component { return &rlcomponents.TurnTakenComponent{} })
	jsonFactory.RegisterComponent("SizeComponent", func() ecs.Component { return &rlcomponents.SizeComponent{} })
	jsonFactory.RegisterComponent("BodyComponent", func() ecs.Component { return &rlcomponents.BodyComponent{} })
	jsonFactory.RegisterComponent("BodyInventoryComponent", func() ecs.Component { return &rlcomponents.BodyInventoryComponent{} })
}
