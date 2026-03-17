package component

import (
	"github.com/mechanical-lich/game-engine/config"
)

// MyTurnComponent .
type AppearanceComponent struct {
	SpriteX      int
	SpriteY      int
	FrameCount   int
	CurrentFrame int
	Resource     string
	R, G, B      uint8
}

func (pc AppearanceComponent) GetType() string {
	return "AppearanceComponent"
}

func (pc AppearanceComponent) GetFrameX() int {
	return pc.SpriteX + (config.TileSizeW * pc.CurrentFrame)
}

func (pc *AppearanceComponent) Update() {
	pc.CurrentFrame++
	if pc.CurrentFrame >= pc.FrameCount {
		pc.CurrentFrame = 0
	}

}
