package system

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/action"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath"
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
	case "rush":
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
	case "face_north":
		rlentity.Face(entity, 0, -1)
		message.AddMessage("Facing north.")
		return nil
	case "face_south":
		rlentity.Face(entity, 0, 1)
		message.AddMessage("Facing south.")
		return nil
	case "face_west":
		rlentity.Face(entity, -1, 0)
		message.AddMessage("Facing west.")
		return nil
	case "face_east":
		rlentity.Face(entity, 1, 0)
		message.AddMessage("Facing east.")
		return nil
	}

	var act action.Action
	switch command {
	case "move_north":
		act = action.MoveAction{DeltaX: 0, DeltaY: -1}
	case "move_south":
		act = action.MoveAction{DeltaX: 0, DeltaY: 1}
	case "move_west":
		act = action.MoveAction{DeltaX: -1, DeltaY: 0}
	case "move_east":
		act = action.MoveAction{DeltaX: 1, DeltaY: 0}
	case "move_northwest":
		act = action.MoveAction{DeltaX: -1, DeltaY: -1}
	case "move_northeast":
		act = action.MoveAction{DeltaX: 1, DeltaY: -1}
	case "move_southwest":
		act = action.MoveAction{DeltaX: -1, DeltaY: 1}
	case "move_southeast":
		act = action.MoveAction{DeltaX: 1, DeltaY: 1}
	case "fire":
		act = action.ShootAction{Aimed: false}
	case "aimed_shot":
		pc := entity.GetComponent("PlayerComponent").(*component.PlayerComponent)
		bodyPart := pc.PendingAimedBodyPart
		pc.PendingAimedBodyPart = ""
		act = action.ShootAction{Aimed: true, AimedBodyPart: bodyPart}
	case "burst_fire":
		act = action.ShootAction{Burst: true}
	case "mouse_shoot":
		pc := entity.GetComponent("PlayerComponent").(*component.PlayerComponent)
		d := pc.PendingMouseShoot
		pc.PendingMouseShoot = nil
		if d != nil {
			epc := entity.GetComponent(component.Position).(*component.PositionComponent)
			if entity.HasComponent(component.Direction) {
				dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
				dc.Direction = rlmath.BestFacingDirection(epc.GetX(), epc.GetY(), d.TargetX, d.TargetY)
			}
			act = action.ShootAction{
				Aimed: d.Aimed, Burst: d.Burst, AimedBodyPart: d.AimedBodyPart,
				TargetX: d.TargetX, TargetY: d.TargetY, HasTarget: true,
			}
		}
	case "heal":
		act = action.HealAction{}
	case "stairs":
		act = action.StairsAction{}
	case "equip":
		act = action.EquipAction{}
	case "pickup":
		act = action.PickupAction{}
	case "pickup_item":
		pc := entity.GetComponent("PlayerComponent").(*component.PlayerComponent)
		d := pc.PendingPickup
		pc.PendingPickup = nil
		if d != nil {
			act = action.PickupItemAction{Item: d.Item, TileX: d.TileX, TileY: d.TileY, TileZ: d.TileZ}
		}
	case "equip_item":
		pc := entity.GetComponent("PlayerComponent").(*component.PlayerComponent)
		d := pc.PendingEquip
		pc.PendingEquip = nil
		if d != nil {
			act = action.EquipItemAction{Item: d.Item, TileX: d.TileX, TileY: d.TileY, TileZ: d.TileZ}
		}
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
