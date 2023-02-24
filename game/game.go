package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/spaceplant/config"
)

type Game struct {
	title string
}

func NewGame(title string) (*Game, error) {
	game := &Game{title: title}
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(title)

	return game, nil
}

func (g *Game) Run() error {
	err := ebiten.RunGame(g)
	return err
}

func (g *Game) Update() error {

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
