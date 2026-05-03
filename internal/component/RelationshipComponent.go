package component

import "github.com/mechanical-lich/mlge/ecs"

// Bond describes a social tie between two crew members.
type Bond struct {
	PartnerName string // DescriptionComponent.Name of the other party
	Type        string // "spouse" | "friend" | "coworker"
}

// RelationshipComponent holds all of a crew member's social bonds.
// When a bonded partner dies, the death listener sets "_bond_broken" in the
// entity's ScriptComponent.Vars so the crew_ai script can react next turn.
type RelationshipComponent struct {
	Bonds []Bond
}

func (c *RelationshipComponent) GetType() ecs.ComponentType { return Relationship }
