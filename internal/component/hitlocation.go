package component

import "github.com/mechanical-lich/mlge/ecs"

// HitLocationComponent defines weighted probabilities for body part targeting.
// Keys match the part names in the entity's BodyComponent.Parts map.
// Higher weight = more likely to be hit. Falls back to equal weight if absent.
type HitLocationComponent struct {
	Weights map[string]int
}

func (h HitLocationComponent) GetType() ecs.ComponentType {
	return HitLocation
}
