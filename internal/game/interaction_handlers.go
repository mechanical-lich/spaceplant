package game

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// interactionListener handles InteractionEvents fired by rlentity.CheckInteraction.
// Register it once in NewMainSimState.
type interactionListener struct {
	sim *SimWorld
}

func (l *interactionListener) HandleEvent(data event.EventData) error {
	ev := data.(rlcomponents.InteractionEvent)
	switch ev.Trigger.Type {

	case "post_message":
		msg := ev.Trigger.Params["message"]
		sender := ev.Trigger.Params["sender"]
		if sender == "" {
			sender = rlentity.GetName(ev.Target)
		}
		message.PostTaggedMessage("interaction", sender, msg)

	case "unlock_door":
		if id := ev.Trigger.Params["target_id"]; id != "" {
			door := rlentity.FindByID(l.sim.Level.Level, id)
			if door != nil && door.HasComponent(rlcomponents.Door) {
				dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
				dc.Open = true
				dc.OwnedBy = ""
			}
		}
		if tag := ev.Trigger.Params["tag"]; tag != "" {
			for _, door := range rlentity.FindByTag(l.sim.Level.Level, tag) {
				if door.HasComponent(rlcomponents.Door) {
					dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
					dc.Open = true
					dc.OwnedBy = ""
				}
			}
		}

	case "lock_door":
		faction := ev.Trigger.Params["faction"]
		if id := ev.Trigger.Params["target_id"]; id != "" {
			door := rlentity.FindByID(l.sim.Level.Level, id)
			if door != nil && door.HasComponent(rlcomponents.Door) {
				dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
				dc.Open = false
				dc.OwnedBy = faction
			}
		}
		if tag := ev.Trigger.Params["tag"]; tag != "" {
			for _, door := range rlentity.FindByTag(l.sim.Level.Level, tag) {
				if door.HasComponent(rlcomponents.Door) {
					dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
					dc.Open = false
					dc.OwnedBy = faction
				}
			}
		}
	}

	return nil
}
