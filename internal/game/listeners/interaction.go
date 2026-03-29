package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// InteractionListener handles InteractionEvents fired by rlentity.CheckInteraction.
type InteractionListener struct {
	Sim SimAccess
}

func (l *InteractionListener) HandleEvent(data mlgeevent.EventData) error {
	ev := data.(rlcomponents.InteractionEvent)

	// Prompt events have no trigger — only show them to the player.
	if ev.Prompt != "" {
		player := l.Sim.GetPlayer()
		if player != nil && ev.Actor == player {
			message.AddMessage(rlentity.GetName(ev.Target) + ": " + ev.Prompt)
		}
		return nil
	}

	switch ev.Trigger.Type {

	case "post_message":
		msg := ev.Trigger.Params["message"]
		sender := ev.Trigger.Params["sender"]
		if sender == "" {
			sender = rlentity.GetName(ev.Target)
		}
		message.PostTaggedMessage("interaction", sender, msg)

	case "unlock_door":
		level := l.Sim.GetRLLevel()
		if id := ev.Trigger.Params["target_id"]; id != "" {
			door := rlentity.FindByID(level, id)
			if door != nil && door.HasComponent(rlcomponents.Door) {
				dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
				dc.Open = true
				dc.OwnedBy = ""
			}
		}
		if tag := ev.Trigger.Params["tag"]; tag != "" {
			for _, door := range rlentity.FindByTag(level, tag) {
				if door.HasComponent(rlcomponents.Door) {
					dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
					dc.Open = true
					dc.OwnedBy = ""
				}
			}
		}

	case "lock_door":
		level := l.Sim.GetRLLevel()
		faction := ev.Trigger.Params["faction"]
		if id := ev.Trigger.Params["target_id"]; id != "" {
			door := rlentity.FindByID(level, id)
			if door != nil && door.HasComponent(rlcomponents.Door) {
				dc := door.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
				dc.Open = false
				dc.OwnedBy = faction
			}
		}
		if tag := ev.Trigger.Params["tag"]; tag != "" {
			for _, door := range rlentity.FindByTag(level, tag) {
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
