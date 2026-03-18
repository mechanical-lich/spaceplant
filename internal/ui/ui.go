package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/state"
)

type GUIViewInterface interface {
	Update(state state.StateInterface)
	Draw(screen *ebiten.Image, s state.StateInterface)
}

type GUIViewBase struct {
}

type GUI struct {
	State GUIViewInterface
}

func NewGUI(startingView GUIViewInterface) *GUI {
	return &GUI{State: startingView}
}

func (g *GUI) Update(s state.StateInterface) {
	if g.State != nil {
		g.State.Update(s)
	}
}

func (g *GUI) Draw(screen *ebiten.Image, s state.StateInterface) {
	g.State.Draw(screen, s)
}
