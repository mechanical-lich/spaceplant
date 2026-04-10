package listeners

import (
	"fmt"

	spcombat "github.com/mechanical-lich/spaceplant/internal/combat"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

// CombatListener handles CombatEvents and logs messages only when the player
// is the attacker or defender. NPC-vs-NPC combat is silently ignored.
type CombatListener struct {
	Sim SimAccess
}

func (l *CombatListener) HandleEvent(evt mlgeevent.EventData) error {
	e, ok := evt.(spcombat.CombatEvent)
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
	if attacker == "" {
		attacker = "Something"
	}
	if playerIsAttacker {
		attacker = "You"
	}

	defender := e.DefenderName
	if defender == "" {
		defender = "something"
	}
	if playerIsDefender {
		defender = "Player"
	}

	source := e.Source
	if source == "" {
		source = "fist"
	}

	var msg string
	switch {
	// Broken/amputated — standalone event fired separately from the damage event.
	case (e.Broken || e.Amputated) && e.Damage == 0:
		if e.Amputated {
			msg = fmt.Sprintf("%s's %s was amputated!", defender, e.BodyPart)
		} else {
			msg = fmt.Sprintf("%s's %s was broken!", defender, e.BodyPart)
		}

	// Save results.
	case e.SavePass:
		if e.Attacker != nil {
			msg = fmt.Sprintf("%s saved against %s's %s", defender, attacker, source)
		} else {
			msg = fmt.Sprintf("%s saved against %s", defender, source)
		}

	case e.SaveFail:
		if e.Attacker != nil {
			if e.Damage > 0 {
				msg = fmt.Sprintf("%s failed save against %s's %s and took %d %s damage!", defender, attacker, source, e.Damage, e.DamageType)
			} else {
				msg = fmt.Sprintf("%s failed save against %s's %s", defender, attacker, source)
			}
		} else {
			if e.Damage > 0 {
				msg = fmt.Sprintf("%s failed save and took %d %s damage!", defender, e.Damage, e.DamageType)
			} else {
				msg = fmt.Sprintf("%s failed save against %s", defender, source)
			}
		}

	// Miss.
	case e.Miss:
		if playerIsAttacker {
			msg = fmt.Sprintf("You missed %s with your %s", defender, source)
		} else {
			msg = fmt.Sprintf("%s missed %s with their %s", attacker, defender, source)
		}
		if playerIsAttacker && e.WoundPenalty > 0 {
			msg += fmt.Sprintf(" (wounded: -%d)", e.WoundPenalty)
		}

	// Normal hit / crit.
	default:
		part := e.BodyPart
		if part == "" {
			part = "body"
		}
		if e.Parried && e.Damage == 0 {
			if playerIsDefender {
				msg = fmt.Sprintf("You parried %s's %s!", attacker, source)
			} else {
				msg = fmt.Sprintf("%s parried %s's %s!", defender, attacker, source)
			}
		} else if e.Parried {
			if playerIsDefender {
				msg = fmt.Sprintf("You partially blocked %s's %s — took %d %s damage!", attacker, source, e.Damage, e.DamageType)
			} else {
				msg = fmt.Sprintf("%s partially blocked %s's %s — took %d %s damage!", defender, attacker, source, e.Damage, e.DamageType)
			}
		} else {
			verb := "hit"
			if e.Crit {
				verb = "critically hit"
			}
			if playerIsAttacker {
				msg = fmt.Sprintf("You %s %s's %s with your %s for %d %s damage!", verb, defender, part, source, e.Damage, e.DamageType)
			} else {
				msg = fmt.Sprintf("%s %s %s's %s with their %s for %d %s damage!", attacker, verb, defender, part, source, e.Damage, e.DamageType)
			}
			if playerIsAttacker && e.WoundPenalty > 0 {
				msg += fmt.Sprintf(" (wounded: -%d)", e.WoundPenalty)
			}
		}
	}

	if msg != "" {
		message.AddMessage(msg)
	}
	return nil
}
