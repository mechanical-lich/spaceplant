package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// BondDeathListener scans all level entities for social bonds to a dying crew
// member and marks survivors via "_bond_broken" in their ScriptComponent.Vars.
// The crew_ai.basic script reads this flag next turn and applies a sanity penalty.
type BondDeathListener struct {
	Sim      SimAccess
	seen     map[string]bool // dying entity names already processed this event cycle
}

func (l *BondDeathListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(rlcomponents.DeathEvent)
	if !ok {
		return nil
	}
	if e.Dying == nil || !e.Dying.HasComponent(component.Description) {
		return nil
	}
	dyingName := e.Dying.GetComponent(component.Description).(*component.DescriptionComponent).Name
	if dyingName == "" {
		return nil
	}

	// Only process each dying entity once even if multiple watchers fire events.
	if l.seen == nil {
		l.seen = make(map[string]bool)
	}
	if l.seen[dyingName] {
		return nil
	}
	l.seen[dyingName] = true

	level := l.Sim.GetRLLevel()
	if level == nil {
		return nil
	}

	for _, entity := range level.Entities {
		if entity == nil || entity == e.Dying {
			continue
		}
		if !entity.HasComponent(component.Relationship) {
			continue
		}
		rc := entity.GetComponent(component.Relationship).(*component.RelationshipComponent)
		for _, bond := range rc.Bonds {
			if bond.PartnerName == dyingName {
				// Flag it in the script vars so crew_ai.basic picks it up next turn.
				if entity.HasComponent(component.Script) {
					sc := entity.GetComponent(component.Script).(*component.ScriptComponent)
					if sc.Vars == nil {
						sc.Vars = make(map[string]any)
					}
					sc.Vars["_bond_broken"] = bond.Type
				}
				// Also update PersonalityComponent.Sanity directly for immediate effect.
				if entity.HasComponent(component.Personality) {
					pc := entity.GetComponent(component.Personality).(*component.PersonalityComponent)
					penalty := 10
					if bond.Type == "spouse" {
						penalty = 30
					}
					pc.Sanity -= penalty
					if pc.Sanity < 0 {
						pc.Sanity = 0
					}
				}
				break
			}
		}
	}
	return nil
}
