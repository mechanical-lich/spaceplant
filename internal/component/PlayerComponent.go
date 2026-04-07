package component

import "github.com/mechanical-lich/mlge/ecs"

// PendingReloadData holds a weapon/ammo pair queued for reload by the UI.
// It is set by ProcessCommand under the sim lock and consumed by PlayerSystem.
type PendingReloadData struct {
	WeaponItem *ecs.Entity
	AmmoItem   *ecs.Entity
}

// PlayerComponent - Handles websocket communications
type PlayerComponent struct {
	Commands             []string
	Rushing              bool
	PendingReload        *PendingReloadData
	PendingAimedBodyPart string // body part chosen in the targeted aimed shot modal
}

// GetType get the type
func (PlayerComponent) GetType() ecs.ComponentType {
	return "PlayerComponent"
}

func (pc *PlayerComponent) PushCommand(x string) {
	pc.Commands = append(pc.Commands, x)
}

func (pc *PlayerComponent) PopCommand() string {
	x := ""
	if len(pc.Commands) > 0 {
		x, pc.Commands = pc.Commands[0], pc.Commands[1:]
	}

	return x
}
