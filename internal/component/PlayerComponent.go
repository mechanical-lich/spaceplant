package component

import "github.com/mechanical-lich/mlge/ecs"

// PlayerComponent - Handles websocket communications
type PlayerComponent struct {
	Commands []string
	Rushing  bool
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
