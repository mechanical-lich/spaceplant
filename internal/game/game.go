package game

import (
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// LoadData loads blueprints, tile definitions, and image assets from disk.
// Call once at startup before creating the simulation world.
func LoadData() error {
	if err := factory.FactoryLoad("data/entity_blueprints.json"); err != nil {
		return err
	}
	if err := world.LoadTileDefinitions("data/tile_definitions.json"); err != nil {
		return err
	}
	return LoadAssets()
}

// LoadDataHeadless loads blueprints and tile definitions without image assets.
// Use this for terminal or headless entry points that have no Ebiten window.
func LoadDataHeadless() error {
	if err := factory.FactoryLoad("data/entity_blueprints.json"); err != nil {
		return err
	}
	return world.LoadTileDefinitions("data/tile_definitions.json")
}

// LoadAssets loads all image textures into the resource cache.
func LoadAssets() error {
	textures := []struct{ key, path string }{
		{"map", "assets/map32x48.png"},
		{"entities", "assets/entities 32x48.png"},
		{"large_entities", "assets/large_entities.png"},
		{"decorations", "assets/decorations.png"},
		{"pickups", "assets/pickups.png"},
		{"fx", "assets/fx.png"},
		{"inventory", "assets/inventory.png"},
	}
	for _, t := range textures {
		if err := resource.LoadImageAsTexture(t.key, t.path); err != nil {
			return err
		}
	}
	return nil
}
