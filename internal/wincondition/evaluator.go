package wincondition

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
)

// EvalContext carries all sim state the evaluator needs.
type EvalContext struct {
	Player            *ecs.Entity
	Entities          []*ecs.Entity // live (non-dead) entities
	Flags             map[string]any
	MotherPlantPlaced bool
}

// Evaluator holds the loaded rule set and evaluates it against an EvalContext.
type Evaluator struct {
	rules []Rule
}

// New creates an Evaluator from the given RuleSet.
func New(rs RuleSet) *Evaluator {
	return &Evaluator{rules: rs.Rules}
}

// EvalKill evaluates all "kill" rules for the given dying blueprint.
// Returns the first matching rule.
func (ev *Evaluator) EvalKill(blueprint string, ctx EvalContext) (Rule, bool) {
	for _, r := range ev.rules {
		if r.Trigger != TriggerKill || r.Blueprint != blueprint {
			continue
		}
		if matchConditions(r.When, ctx) {
			return r, true
		}
	}
	return Rule{}, false
}

// EvalInteraction evaluates all "interaction" rules for the given interaction name.
func (ev *Evaluator) EvalInteraction(name string, ctx EvalContext) (Rule, bool) {
	for _, r := range ev.rules {
		if r.Trigger != TriggerInteraction || r.Interaction != name {
			continue
		}
		if matchConditions(r.When, ctx) {
			return r, true
		}
	}
	return Rule{}, false
}

// EvalPlayerDeath evaluates all "player_death" rules.
func (ev *Evaluator) EvalPlayerDeath(ctx EvalContext) (Rule, bool) {
	for _, r := range ev.rules {
		if r.Trigger != TriggerPlayerDeath {
			continue
		}
		if matchConditions(r.When, ctx) {
			return r, true
		}
	}
	return Rule{}, false
}

func matchConditions(when []Condition, ctx EvalContext) bool {
	for _, c := range when {
		if !matchCondition(c, ctx) {
			return false
		}
	}
	return true
}

func matchCondition(c Condition, ctx EvalContext) bool {
	if c.PlayerBackground != nil {
		if ctx.Player == nil {
			return false
		}
		bg, ok := ctx.Player.GetComponent(component.Background).(*component.BackgroundComponent)
		if !ok || bg == nil {
			return false
		}
		if bg.BackgroundID != *c.PlayerBackground {
			return false
		}
	}

	if c.PlayerClass != nil {
		if ctx.Player == nil {
			return false
		}
		cls, ok := ctx.Player.GetComponent(component.Class).(*component.ClassComponent)
		if !ok || cls == nil {
			return false
		}
		if !cls.HasClass(*c.PlayerClass) {
			return false
		}
	}

	if c.GameFlag != nil {
		switch *c.GameFlag {
		case "mother_plant_placed":
			if !ctx.MotherPlantPlaced {
				return false
			}
		default:
			v := ctx.Flags[*c.GameFlag]
			if v == nil || v == false || v == 0.0 {
				return false
			}
		}
	}

	if c.EntityCount != nil {
		ec := c.EntityCount
		count := 0
		for _, e := range ctx.Entities {
			if e != nil && e.Blueprint == ec.Blueprint {
				count++
			}
		}
		if !applyOp(ec.Op, count, ec.Value) {
			return false
		}
	}

	return true
}

// FireRule emits GameWonEventData or GameLostEventData based on rule.Result.
// Detail is the raw death cause (passed through to GameLostEventData).
func FireRule(rule Rule, detail string) {
	switch rule.Result {
	case ResultWin:
		eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{
			Outcome: rule.Outcome,
			Message: rule.Message,
		})
	case ResultLose:
		eventsystem.EventManager.SendEvent(eventsystem.GameLostEventData{
			Outcome: rule.Outcome,
			Message: rule.Message,
			Detail:  detail,
		})
	}
}

func applyOp(op string, a, b int) bool {
	switch op {
	case "eq":
		return a == b
	case "gt":
		return a > b
	case "lt":
		return a < b
	case "gte":
		return a >= b
	case "lte":
		return a <= b
	default:
		return false
	}
}
