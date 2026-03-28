package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

type PlayerSystem struct {
}

func (s *PlayerSystem) UpdateSystem(data any) error {
	return nil
}

func (s *PlayerSystem) Requires() []ecs.ComponentType {
	return nil
}

// UpdateEntity processes one player action per tick when the player has MyTurn.
func (s *PlayerSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	if !entity.HasComponent(component.Player) {
		return nil
	}
	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}
	if !entity.HasComponent(rlcomponents.MyTurn) {
		return nil
	}

	l := levelInterface.(*world.Level)
	playerComponent := entity.GetComponent(component.Player).(*component.PlayerComponent)
	command := playerComponent.PopCommand()

	// Non-turn-consuming commands.
	switch command {
	case "R":
		ec := entity.GetComponent(component.Energy).(*component.EnergyComponent)
		if playerComponent.Rushing {
			ec.Speed /= 2
			playerComponent.Rushing = false
			message.AddMessage("Rush mode off.")
		} else {
			ec.Speed *= 2
			playerComponent.Rushing = true
			message.AddMessage("Rush mode on!")
		}
		return nil
	}

	var act action.Action
	switch command {
	case "W":
		act = action.MoveAction{DeltaX: 0, DeltaY: -1}
	case "S":
		act = action.MoveAction{DeltaX: 0, DeltaY: 1}
	case "A":
		act = action.MoveAction{DeltaX: -1, DeltaY: 0}
	case "D":
		act = action.MoveAction{DeltaX: 1, DeltaY: 0}
	case "F":
		dir := playerComponent.PopCommand()
		act = action.ShootAction{Direction: dir}
	case "H":
		act = action.HealAction{}
	case "Period":
		act = action.StairsAction{}
	case "E":
		act = action.EquipAction{}
	case "P":
		act = action.PickupAction{}
	default:
		// Check if a skill provides a binding for this key.
		act = skill.ActionForKey(entity, command)
	}

	if act == nil {
		return nil
	}

	entity.AddComponent(rlcomponents.GetTurnTaken())
	err := act.Execute(entity, l)

	// Re-sync item-granted skills after any equip operation.
	if _, isEquip := act.(action.EquipAction); isEquip {
		skill.SyncEquippedSkills(entity)
	}

	return err
}
