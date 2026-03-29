package component

import "github.com/mechanical-lich/mlge/ecs"

// BackgroundComponent records which background the player chose at character creation.
type BackgroundComponent struct {
	BackgroundID string
}

func (c *BackgroundComponent) GetType() ecs.ComponentType { return Background }
