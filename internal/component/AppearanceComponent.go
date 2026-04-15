package component

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent .
type AppearanceComponent struct {
	SpriteX       int
	SpriteY       int
	SpriteHeight  int
	SpriteWidth   int
	SpriteOffsetX int
	SpriteOffsetY int
	FrameCount    int
	CurrentFrame  int
	Resource      string
	R, G, B       uint8
}

func (pc AppearanceComponent) GetType() ecs.ComponentType {
	return "AppearanceComponent"
}

func (pc AppearanceComponent) GetFrameX() int {
	return pc.SpriteX + (pc.SpriteWidth * pc.CurrentFrame)
}

func (pc *AppearanceComponent) Update() {
	pc.CurrentFrame++
	if pc.CurrentFrame >= pc.FrameCount {
		pc.CurrentFrame = 0
	}
}

// SetSprite implements rlsystems.AppearanceUpdater, used by DoorSystem.
func (pc *AppearanceComponent) SetSprite(x, y int) {
	pc.SpriteX = x
	pc.SpriteY = y
}
