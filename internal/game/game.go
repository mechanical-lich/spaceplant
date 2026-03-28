package game

import (
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// LoadData loads blueprints, tile definitions, and image assets from disk.
// Call once at startup before creating the simulation world.
func LoadData() error {
	if err := factory.FactoryLoad("data/blueprints"); err != nil {
		return err
	}
	if err := world.LoadTileDefinitions("data/tile_definitions.json"); err != nil {
		return err
	}
	if err := skill.Load("data/skills/skills.json"); err != nil {
		return err
	}
	return LoadAssets()
}

// LoadDataHeadless loads blueprints and tile definitions without image assets.
// Use this for terminal or headless entry points that have no Ebiten window.
func LoadDataHeadless() error {
	if err := factory.FactoryLoad("data/blueprints"); err != nil {
		return err
	}
	if err := skill.Load("data/skills/skills.json"); err != nil {
		return err
	}
	return world.LoadTileDefinitions("data/tile_definitions.json")
}

// LoadAssets loads all image textures into the resource cache.
func LoadAssets() error {
	return resource.LoadAssetsFromJSON("data/assets.json")
}
