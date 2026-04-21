package wincondition

import (
	"testing"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// helpers

func ptr[T any](v T) *T { return &v }

func newPlayer(backgroundID, classID string) *ecs.Entity {
	e := &ecs.Entity{}
	if backgroundID != "" {
		e.AddComponent(&component.BackgroundComponent{BackgroundID: backgroundID})
	}
	if classID != "" {
		e.AddComponent(&component.ClassComponent{Classes: []string{classID}})
	}
	return e
}

func newEntity(blueprint string) *ecs.Entity {
	return &ecs.Entity{Blueprint: blueprint}
}

// EvalKill

func TestEvalKill_Match(t *testing.T) {
	ev := New(RuleSet{Rules: []Rule{
		{ID: "extermination", Trigger: TriggerKill, Blueprint: "mother_plant",
			Result: ResultWin, Outcome: "extermination"},
	}})
	rule, ok := ev.EvalKill("mother_plant", EvalContext{})
	if !ok {
		t.Fatal("expected match")
	}
	if rule.Outcome != "extermination" {
		t.Errorf("outcome = %q, want extermination", rule.Outcome)
	}
}

func TestEvalKill_WrongBlueprint(t *testing.T) {
	ev := New(RuleSet{Rules: []Rule{
		{ID: "extermination", Trigger: TriggerKill, Blueprint: "mother_plant", Result: ResultWin},
	}})
	_, ok := ev.EvalKill("captain", EvalContext{})
	if ok {
		t.Fatal("expected no match for wrong blueprint")
	}
}

func TestEvalKill_NoRules(t *testing.T) {
	ev := New(RuleSet{})
	_, ok := ev.EvalKill("mother_plant", EvalContext{})
	if ok {
		t.Fatal("expected no match with empty rule set")
	}
}

// EvalInteraction

func TestEvalInteraction_DefaultEscape(t *testing.T) {
	ev := New(RuleSet{Rules: []Rule{
		{ID: "escape_selfish", Trigger: TriggerInteraction, Interaction: "life_pod_escape",
			Result: ResultWin, Outcome: "escape_selfish"},
	}})
	rule, ok := ev.EvalInteraction("life_pod_escape", EvalContext{})
	if !ok {
		t.Fatal("expected match")
	}
	if rule.Outcome != "escape_selfish" {
		t.Errorf("outcome = %q, want escape_selfish", rule.Outcome)
	}
}

func TestEvalInteraction_FirstMatchWins(t *testing.T) {
	player := newPlayer("saboteur", "")
	ev := New(RuleSet{Rules: []Rule{
		{ID: "saboteur_escape", Trigger: TriggerInteraction, Interaction: "life_pod_escape",
			When:   []Condition{{PlayerBackground: ptr("saboteur")}, {GameFlag: ptr("mother_plant_placed")}},
			Result: ResultWin, Outcome: "saboteur"},
		{ID: "escape_selfish", Trigger: TriggerInteraction, Interaction: "life_pod_escape",
			Result: ResultWin, Outcome: "escape_selfish"},
	}})

	// saboteur with plant placed → first rule wins
	rule, ok := ev.EvalInteraction("life_pod_escape", EvalContext{
		Player:            player,
		MotherPlantPlaced: true,
	})
	if !ok || rule.Outcome != "saboteur" {
		t.Errorf("expected saboteur, got %q ok=%v", rule.Outcome, ok)
	}

	// saboteur without plant → first rule fails, falls through to escape_selfish
	rule, ok = ev.EvalInteraction("life_pod_escape", EvalContext{
		Player:            player,
		MotherPlantPlaced: false,
	})
	if !ok || rule.Outcome != "escape_selfish" {
		t.Errorf("expected escape_selfish, got %q ok=%v", rule.Outcome, ok)
	}
}

func TestEvalInteraction_SaboteurCaught(t *testing.T) {
	player := newPlayer("saboteur", "")
	ev := New(RuleSet{Rules: []Rule{
		{ID: "saboteur_caught", Trigger: TriggerInteraction, Interaction: "life_pod_escape",
			When: []Condition{
				{PlayerBackground: ptr("saboteur")},
				{EntityCount: &EntityCount{Blueprint: "mother_plant", Op: "eq", Value: 0}},
			},
			Result: ResultLose, Outcome: "saboteur_caught"},
		{ID: "escape_selfish", Trigger: TriggerInteraction, Interaction: "life_pod_escape",
			Result: ResultWin, Outcome: "escape_selfish"},
	}})

	// saboteur, no mother plant alive → caught
	rule, ok := ev.EvalInteraction("life_pod_escape", EvalContext{
		Player:   player,
		Entities: []*ecs.Entity{},
	})
	if !ok || rule.Outcome != "saboteur_caught" {
		t.Errorf("expected saboteur_caught, got %q ok=%v", rule.Outcome, ok)
	}

	// saboteur, mother plant alive → not caught, falls through to escape_selfish
	rule, ok = ev.EvalInteraction("life_pod_escape", EvalContext{
		Player:   player,
		Entities: []*ecs.Entity{newEntity("mother_plant")},
	})
	if !ok || rule.Outcome != "escape_selfish" {
		t.Errorf("expected escape_selfish, got %q ok=%v", rule.Outcome, ok)
	}
}

// EvalPlayerDeath

func TestEvalPlayerDeath_HeroicDeathBeforeDefault(t *testing.T) {
	ev := New(RuleSet{Rules: []Rule{
		{ID: "heroic_death", Trigger: TriggerPlayerDeath,
			When:   []Condition{{GameFlag: ptr("self_destruct_armed")}},
			Result: ResultWin, Outcome: "heroic_death"},
		{ID: "player_death_default", Trigger: TriggerPlayerDeath,
			Result: ResultLose, Outcome: "dead"},
	}})

	// self-destruct armed → heroic win
	rule, ok := ev.EvalPlayerDeath(EvalContext{SelfDestructArmed: true})
	if !ok || rule.Outcome != "heroic_death" {
		t.Errorf("expected heroic_death, got %q ok=%v", rule.Outcome, ok)
	}

	// not armed → default lose
	rule, ok = ev.EvalPlayerDeath(EvalContext{SelfDestructArmed: false})
	if !ok || rule.Outcome != "dead" {
		t.Errorf("expected dead, got %q ok=%v", rule.Outcome, ok)
	}
}

func TestEvalPlayerDeath_NoRules(t *testing.T) {
	ev := New(RuleSet{})
	_, ok := ev.EvalPlayerDeath(EvalContext{})
	if ok {
		t.Fatal("expected no match with empty rule set")
	}
}

// Condition: PlayerBackground

func TestCondition_PlayerBackground_Match(t *testing.T) {
	player := newPlayer("saboteur", "")
	c := Condition{PlayerBackground: ptr("saboteur")}
	if !matchCondition(c, EvalContext{Player: player}) {
		t.Error("expected match")
	}
}

func TestCondition_PlayerBackground_Mismatch(t *testing.T) {
	player := newPlayer("scientist", "")
	c := Condition{PlayerBackground: ptr("saboteur")}
	if matchCondition(c, EvalContext{Player: player}) {
		t.Error("expected no match")
	}
}

func TestCondition_PlayerBackground_NilPlayer(t *testing.T) {
	c := Condition{PlayerBackground: ptr("saboteur")}
	if matchCondition(c, EvalContext{Player: nil}) {
		t.Error("expected no match with nil player")
	}
}

func TestCondition_PlayerBackground_NoComponent(t *testing.T) {
	player := &ecs.Entity{}
	c := Condition{PlayerBackground: ptr("saboteur")}
	if matchCondition(c, EvalContext{Player: player}) {
		t.Error("expected no match without background component")
	}
}

// Condition: PlayerClass

func TestCondition_PlayerClass_Match(t *testing.T) {
	player := newPlayer("", "security")
	c := Condition{PlayerClass: ptr("security")}
	if !matchCondition(c, EvalContext{Player: player}) {
		t.Error("expected match")
	}
}

func TestCondition_PlayerClass_Mismatch(t *testing.T) {
	player := newPlayer("", "scientist")
	c := Condition{PlayerClass: ptr("security")}
	if matchCondition(c, EvalContext{Player: player}) {
		t.Error("expected no match")
	}
}

// Condition: GameFlag

func TestCondition_GameFlag_SelfDestructArmed(t *testing.T) {
	c := Condition{GameFlag: ptr("self_destruct_armed")}
	if !matchCondition(c, EvalContext{SelfDestructArmed: true}) {
		t.Error("expected match when armed")
	}
	if matchCondition(c, EvalContext{SelfDestructArmed: false}) {
		t.Error("expected no match when not armed")
	}
}

func TestCondition_GameFlag_MotherPlantPlaced(t *testing.T) {
	c := Condition{GameFlag: ptr("mother_plant_placed")}
	if !matchCondition(c, EvalContext{MotherPlantPlaced: true}) {
		t.Error("expected match when placed")
	}
	if matchCondition(c, EvalContext{MotherPlantPlaced: false}) {
		t.Error("expected no match when not placed")
	}
}

func TestCondition_GameFlag_Unknown(t *testing.T) {
	c := Condition{GameFlag: ptr("nonexistent_flag")}
	if matchCondition(c, EvalContext{}) {
		t.Error("expected no match for unknown flag")
	}
}

// Condition: EntityCount

func TestCondition_EntityCount_Eq(t *testing.T) {
	entities := []*ecs.Entity{newEntity("mother_plant"), newEntity("mother_plant")}
	c := Condition{EntityCount: &EntityCount{Blueprint: "mother_plant", Op: "eq", Value: 2}}
	if !matchCondition(c, EvalContext{Entities: entities}) {
		t.Error("expected eq match")
	}
	c2 := Condition{EntityCount: &EntityCount{Blueprint: "mother_plant", Op: "eq", Value: 1}}
	if matchCondition(c2, EvalContext{Entities: entities}) {
		t.Error("expected eq no match")
	}
}

func TestCondition_EntityCount_Gt(t *testing.T) {
	entities := []*ecs.Entity{newEntity("captain"), newEntity("captain"), newEntity("captain")}
	c := Condition{EntityCount: &EntityCount{Blueprint: "captain", Op: "gt", Value: 2}}
	if !matchCondition(c, EvalContext{Entities: entities}) {
		t.Error("expected gt match")
	}
}

func TestCondition_EntityCount_ZeroEntities(t *testing.T) {
	c := Condition{EntityCount: &EntityCount{Blueprint: "mother_plant", Op: "eq", Value: 0}}
	if !matchCondition(c, EvalContext{Entities: []*ecs.Entity{}}) {
		t.Error("expected match for zero count")
	}
}

func TestCondition_EntityCount_NilInSlice(t *testing.T) {
	entities := []*ecs.Entity{nil, newEntity("mother_plant"), nil}
	c := Condition{EntityCount: &EntityCount{Blueprint: "mother_plant", Op: "eq", Value: 1}}
	if !matchCondition(c, EvalContext{Entities: entities}) {
		t.Error("expected nil entries to be skipped")
	}
}

// applyOp

func TestApplyOp(t *testing.T) {
	cases := []struct {
		op   string
		a, b int
		want bool
	}{
		{"eq", 3, 3, true}, {"eq", 3, 4, false},
		{"gt", 4, 3, true}, {"gt", 3, 3, false},
		{"lt", 2, 3, true}, {"lt", 3, 3, false},
		{"gte", 3, 3, true}, {"gte", 4, 3, true}, {"gte", 2, 3, false},
		{"lte", 3, 3, true}, {"lte", 2, 3, true}, {"lte", 4, 3, false},
		{"bad", 1, 1, false},
	}
	for _, tc := range cases {
		got := applyOp(tc.op, tc.a, tc.b)
		if got != tc.want {
			t.Errorf("applyOp(%q, %d, %d) = %v, want %v", tc.op, tc.a, tc.b, got, tc.want)
		}
	}
}

// Loader

func TestLoadFromRules_PlantzRules(t *testing.T) {
	LoadFromRules(RuleSet{Rules: []Rule{
		{ID: "kill_mother_plant", Trigger: "kill", Blueprint: "mother_plant", Result: ResultWin, Outcome: "extermination"},
		{ID: "life_pod_escape", Trigger: "interaction", Interaction: "life_pod_escape", Result: ResultWin, Outcome: "escape_selfish"},
		{ID: "player_death_default", Trigger: "player_death", Result: ResultLose, Outcome: "dead"},
	}})
	ev := Active()

	// kill mother_plant → extermination win
	rule, ok := ev.EvalKill("mother_plant", EvalContext{})
	if !ok || rule.Outcome != "extermination" || rule.Result != ResultWin {
		t.Errorf("extermination: got %+v ok=%v", rule, ok)
	}

	// life pod escape → escape_selfish win
	rule, ok = ev.EvalInteraction("life_pod_escape", EvalContext{})
	if !ok || rule.Outcome != "escape_selfish" || rule.Result != ResultWin {
		t.Errorf("escape_selfish: got %+v ok=%v", rule, ok)
	}

	// player death default → lose
	rule, ok = ev.EvalPlayerDeath(EvalContext{})
	if !ok || rule.Result != ResultLose {
		t.Errorf("player_death_default: got %+v ok=%v", rule, ok)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if err := Load("nonexistent.json"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	if err := Load("testdata/invalid.json"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
