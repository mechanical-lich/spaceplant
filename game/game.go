package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/mlge/state"
	"github.com/mechanical-lich/spaceplant/config"
	"github.com/mechanical-lich/spaceplant/factory"
	"github.com/mechanical-lich/spaceplant/world"
)

type Game struct {
	title        string
	StateMachine state.StateMachine
}

func NewGame(title string) (*Game, error) {
	game := &Game{title: title}
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(title)
	// Load Blueprints (JSON format)
	err := factory.FactoryLoad("data/entity_blueprints.json")
	if err != nil {
		return nil, err
	}

	err = world.LoadTileDefinitions("data/tile_definitions.json")
	if err != nil {
		return nil, err
	}

	err = LoadAssets()
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
	ebiten.SetWindowTitle(fmt.Sprintf("%s - %d", g.title, int(ebiten.ActualFPS())))
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
	err := resource.LoadImageAsTexture("map", "assets/map32x48.png")
	if err != nil {
		return err
	}

	err = resource.LoadImageAsTexture("entities", "assets/entities 32x48.png")
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

	err = resource.LoadImageAsTexture("inventory", "assets/inventory.png")
	if err != nil {
		return err
	}

	return nil
}
