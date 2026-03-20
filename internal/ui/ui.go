package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type GUIViewInterface interface {
	Update(state any)
	Draw(screen *ebiten.Image, s any)
}

type GUIViewBase struct {
}

type GUI struct {
	State GUIViewInterface
}

func NewGUI(startingView GUIViewInterface) *GUI {
	return &GUI{State: startingView}
}

func (g *GUI) Update(s any) {
	if g.State != nil {
		g.State.Update(s)
	}
}

func (g *GUI) Draw(screen *ebiten.Image, s any) {
	g.State.Draw(screen, s)
}
