package listeners

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/rlbodycombat"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// CombatListener handles CombatEvents and logs messages only when the player
// is the attacker or defender. NPC-vs-NPC combat is silently ignored.
type CombatListener struct {
	Sim SimAccess
}

func (l *CombatListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(rlbodycombat.CombatEvent)
	if !ok {
		return nil
	}
	player := l.Sim.GetPlayer()
	if player == nil {
		return nil
	}

	playerIsAttacker := e.Attacker == player
	playerIsDefender := e.Defender == player
	if !playerIsAttacker && !playerIsDefender {
		return nil
	}

	attacker := e.AttackerName
	if playerIsAttacker {
		attacker = "You"
	}
	defender := e.DefenderName
	if playerIsDefender {
		defender = "Player"
	}

	var msg string
	switch {
	case e.SavePass:
		part := e.BodyPart
		if part == "" {
			part = "body"
		}
		msg = fmt.Sprintf("%s's %s saved against %s", defender, part, e.DamageType)

	case e.SaveFail:
		part := e.BodyPart
		if part == "" {
			part = "body"
		}
		msg = fmt.Sprintf("%s's %s failed to save against %d (%s)", defender, part, e.Damage, e.DamageType)
		if e.Amputated {
			msg += fmt.Sprintf(" — %s's %s was amputated!", defender, part)
		} else if e.Broken {
			msg += fmt.Sprintf(" — %s's %s was broken!", defender, part)
		}

	case e.Miss:
		if playerIsAttacker {
			msg = fmt.Sprintf("You missed %s", defender)
		} else {
			msg = fmt.Sprintf("%s missed you", attacker)
		}

	default:
		verb := "hit"
		if e.Crit {
			verb = "critically hit"
		}
		part := e.BodyPart
		if part == "" {
			part = "body"
		}
		msg = fmt.Sprintf("%s %s %s's %s for %d (%s)", attacker, verb, defender, part, e.Damage, e.DamageType)
		if e.Amputated {
			msg += fmt.Sprintf(" — %s's %s was amputated!", defender, part)
		} else if e.Broken {
			msg += fmt.Sprintf(" — %s's %s was broken!", defender, part)
		}
	}

	message.AddMessage(msg)
	return nil
}
