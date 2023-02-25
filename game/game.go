package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/game-engine/resource"
	"github.com/mechanical-lich/game-engine/state"
	"github.com/mechanical-lich/spaceplant/config"
)

type Game struct {
	title        string
	StateMachine state.StateMachine
}

func NewGame(title string) (*Game, error) {
	game := &Game{title: title}
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(title)

	err := LoadAssets()
	if err != nil {
		return nil, err
	}

	mainState, err := NewMainState()
	if err != nil {
		return nil, err
	}

	game.StateMachine.PushState(mainState)

	return game, nil
}

func (g *Game) Run() error {
	err := ebiten.RunGame(g)
	return err
}

// Main update loop
func (g *Game) Update() error {
	g.StateMachine.Update()
	return nil
}

// Main render loop
func (g *Game) Draw(screen *ebiten.Image) {
	g.StateMachine.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

func LoadAssets() error {
	err := resource.LoadImageAsTexture("map", "assets/map.png")
	if err != nil {
		return err
	}

	err = resource.LoadImageAsTexture("entities", "assets/entities.png")
	if err != nil {
		return err
	}

	err = resource.LoadImageAsTexture("decorations", "assets/decorations.png")
	if err != nil {
		return err
	}

	err = resource.LoadImageAsTexture("pickups", "assets/pickups.png")
	if err != nil {
		return err
	}

	err = resource.LoadImageAsTexture("fx", "assets/fx.png")
	if err != nil {
		return err
	}

	return nil
}
