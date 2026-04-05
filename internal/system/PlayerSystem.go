package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
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

	// Pending reload (queued by the reload modal via CmdReload).
	if playerComponent.PendingReload != nil {
		reload := playerComponent.PendingReload
		playerComponent.PendingReload = nil
		playerComponent.PopCommand() // drain the sentinel pushed to unblock phaseWaitingForInput
		act := action.ReloadAction{WeaponItem: reload.WeaponItem, AmmoItem: reload.AmmoItem}
		entity.AddComponent(rlcomponents.GetTurnTaken())
		return act.Execute(entity, l)
	}

	command := playerComponent.PopCommand()

	// Non-turn-consuming commands.
	switch command {
	case "r":
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
	case "W":
		rlentity.Face(entity, 0, -1)
		message.AddMessage("Facing north.")
		return nil
	case "S":
		rlentity.Face(entity, 0, 1)
		message.AddMessage("Facing south.")
		return nil
	case "A":
		rlentity.Face(entity, -1, 0)
		message.AddMessage("Facing west.")
		return nil
	case "D":
		rlentity.Face(entity, 1, 0)
		message.AddMessage("Facing east.")
		return nil
	}

	var act action.Action
	switch command {
	case "w":
		act = action.MoveAction{DeltaX: 0, DeltaY: -1}
	case "s":
		act = action.MoveAction{DeltaX: 0, DeltaY: 1}
	case "a":
		act = action.MoveAction{DeltaX: -1, DeltaY: 0}
	case "d":
		act = action.MoveAction{DeltaX: 1, DeltaY: 0}
	case "f":
		act = action.ShootAction{Aimed: false}
	case "F":
		act = action.ShootAction{Aimed: true}
	case "g":
		act = action.ShootAction{Burst: true}
	case "h":
		act = action.HealAction{}
	case "Period":
		act = action.StairsAction{}
	case "e":
		act = action.EquipAction{}
	case "p":
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
