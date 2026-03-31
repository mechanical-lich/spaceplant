package component

import "github.com/mechanical-lich/mlge/ecs"

// LayeredAppearanceComponent replaces AppearanceComponent for entities that
// use the composited sprite system. The entity is rendered as stacked layers
// (body, equipped clothing, hair, headwear) sourced from the mid_* / slim_*
// textures loaded from assets/32x48/mid and assets/32x48/slim.
type LayeredAppearanceComponent struct {
	BodyType  string // "mid" or "slim"
	BodyIndex int    // 0-based column on the body sheet (skin tone / style)
	HairIndex int    // 0-based column on the hair sheet; -1 = no hair
	Randomize bool   // if true, factory fills in random BodyType/BodyIndex/HairIndex
}

func (c *LayeredAppearanceComponent) GetType() ecs.ComponentType {
	return LayeredAppearance
}

// WearableAppearanceComponent is added to item entities that have a visible
// layer when equipped. Layer identifies which sprite sheet to composite
// ("shirt", "pants", "shoes", "headwear"). Index is the 0-based column.
type WearableAppearanceComponent struct {
	Layer string // "shirt", "pants", "shoes", "headwear"
	Index int    // column index on the layer sheet (0-based)
}

func (c *WearableAppearanceComponent) GetType() ecs.ComponentType {
	return WearableAppearance
}
