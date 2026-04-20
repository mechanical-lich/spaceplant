# Scenarios — Developer Guide

A scenario is the self-contained "threat" for a station run. It controls which monsters spawn, how aggressively, which win/lose rules apply, and which extra skills and backgrounds appear in character creation. One scenario is randomly chosen (or player-selected) when a new station is generated.

---

## Architecture

| Layer | File |
|-------|------|
| Scenario type & fields | `internal/scenario/types.go` |
| Loading, selection, singleton | `internal/scenario/loader.go` |
| Data files | `data/scenarios/*.json` |
| Wired into startup | `internal/game/game.go` (`LoadData` / `LoadDataHeadless`) |
| Monster spawning | `internal/gamemaster/gm.go` |
| Character creator extras | `internal/game/character_creator.go` |
| New-station picker UI | `internal/game/title_screen.go` |

### Startup flow

```
game.LoadData()
  └── scenario.Load("data/scenarios")      // reads all *.json, filters enabled:true, picks random
  └── wincondition.LoadFromRules(...)       // installs the scenario's win/lose rules as active
```

`scenario.Active()` returns the chosen scenario for the rest of the run. `GameMaster.Init` and `GameMaster.Update` read from it for spawn lists and counts.

---

## JSON Schema — `data/scenarios/*.json`

```json
{
  "id": "my_scenario",
  "name": "Display Name",
  "description": "Shown nowhere yet — useful for dev notes.",
  "enabled": true,

  "hostiles":      ["blueprint_id", ...],
  "rare_hostiles": ["blueprint_id", ...],
  "hostile_initial": 15,
  "hostile_max":    20,
  "spawn_chance":   0.4,

  "extra_skills":      ["skill_id", ...],
  "extra_backgrounds": ["background_id", ...],

  "win_conditions": {
    "rules": [ ... ]
  }
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier used by `scenario.SelectByID`. |
| `name` | string | Shown in the New Station scenario picker. |
| `enabled` | bool | Must be `true` to be included in random selection. Default is `false` (Go zero value). |
| `hostiles` | []string | Blueprint IDs for common spawns. Picked uniformly at random. |
| `rare_hostiles` | []string | Blueprint IDs for rare spawns (1-in-20 chance per spawn tick). |
| `hostile_initial` | int | How many hostiles to place when each floor is initialised. |
| `hostile_max` | int | Per-floor cap on concurrent hostile entities. |
| `spawn_chance` | float64 | Probability (0.0–1.0) that a new hostile spawns each game tick. |
| `extra_skills` | []string | Skill IDs added to the character creation skill list for this scenario. |
| `extra_backgrounds` | []string | Background IDs added to the character creation background list. These must already exist in `data/backgrounds/backgrounds.json`. |
| `win_conditions` | RuleSet | Embedded win/lose rules. Same schema as the old `data/win_conditions/default.json`. See [Win Conditions](win_conditions.md). |

---

## Enabling and Disabling Scenarios

Set `"enabled": false` to exclude a scenario from random selection without deleting it. The loader returns an error if no enabled scenarios are found.

The New Station screen only shows enabled scenarios in its dropdown. If only one scenario is enabled, the dropdown still appears (with "Random" + that one entry).

---

## Selecting a Scenario Programmatically

```go
// Pick randomly from all enabled scenarios (called automatically by Load):
scenario.SelectRandom()

// Force a specific scenario by ID:
scenario.SelectByID("plantz")

// Get the active scenario:
s := scenario.Active()
```

After calling `SelectByID` or `SelectRandom`, re-install the win conditions:

```go
wincondition.LoadFromRules(scenario.Active().WinConditions)
```

`title_screen.go` does this automatically when the player clicks Generate.

---

## Win Conditions in a Scenario

Win/lose rules are embedded directly in the scenario JSON under `"win_conditions"`. The schema is identical to the old standalone `data/win_conditions/default.json` (which is now superseded).

```json
"win_conditions": {
  "rules": [
    {
      "id": "kill_boss",
      "trigger": "kill",
      "blueprint": "my_boss",
      "result": "win",
      "outcome": "killed_boss",
      "message": "The boss is dead."
    },
    {
      "id": "player_death_default",
      "trigger": "player_death",
      "result": "lose",
      "outcome": "dead",
      "message": ""
    }
  ]
}
```

Always include a `player_death` catch-all rule with no `when` guards — without it, dying produces no outcome.

See [Win Conditions](win_conditions.md) for the full trigger/condition reference.

---

## Adding a New Scenario

### 1. Add monster blueprints (if needed)

Create a JSON file under `data/blueprints/entities/<scenario_name>/monsters.json`. Follow the structure of `data/blueprints/entities/plantz/monsters.json`. The blueprint IDs you define here are what you put in `hostiles` and `rare_hostiles`.

### 2. Add scenario-specific skills and backgrounds (if needed)

- Skills → `data/skills/skills.json` (see [Adding Skills](adding_skills.md))
- Backgrounds → `data/backgrounds/backgrounds.json`

Reference their IDs in `extra_skills` / `extra_backgrounds`.

### 3. Create `data/scenarios/<id>.json`

Start with `"enabled": false` while developing. Add your monster blueprint IDs, tuning values, any extra skills/backgrounds, and the `win_conditions` rule set.

### 4. Enable and test

Set `"enabled": true`. On the New Station screen, choose the scenario by name from the dropdown to test it deterministically without waiting for the random pick.

---

## Existing Scenarios

| ID | Name | Status |
|----|------|--------|
| `plantz` | Plant Infestation | Enabled — the original scenario |
| `demon_invasion` | Event Horizon | Disabled — stub, monster blueprints not yet created |
| `xenomorph` | Alien Infestation | Disabled — stub, monster blueprints not yet created |
