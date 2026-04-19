package wincondition

// TriggerType identifies which event fires a rule.
type TriggerType string

const (
	TriggerKill        TriggerType = "kill"
	TriggerInteraction TriggerType = "interaction"
	TriggerPlayerDeath TriggerType = "player_death"
)

// ResultType is either "win" or "lose".
type ResultType string

const (
	ResultWin  ResultType = "win"
	ResultLose ResultType = "lose"
)

// Condition is one optional guard. All guards on a rule must pass.
type Condition struct {
	PlayerBackground *string      `json:"player_background,omitempty"`
	PlayerClass      *string      `json:"player_class,omitempty"`
	GameFlag         *string      `json:"game_flag,omitempty"`
	EntityCount      *EntityCount `json:"entity_count,omitempty"`
}

// EntityCount tests the count of live entities with a given blueprint.
type EntityCount struct {
	Blueprint string `json:"blueprint"`
	Op        string `json:"op"` // "eq","gt","lt","gte","lte"
	Value     int    `json:"value"`
}

// Rule is one complete win/loss entry.
type Rule struct {
	ID          string      `json:"id"`
	Trigger     TriggerType `json:"trigger"`
	Blueprint   string      `json:"blueprint,omitempty"`   // for "kill"
	Interaction string      `json:"interaction,omitempty"` // for "interaction"
	When        []Condition `json:"when,omitempty"`        // all must match
	Result      ResultType  `json:"result"`
	Outcome     string      `json:"outcome"`
	Message     string      `json:"message"`
}

// RuleSet is the top-level JSON document.
type RuleSet struct {
	Rules []Rule `json:"rules"`
}
