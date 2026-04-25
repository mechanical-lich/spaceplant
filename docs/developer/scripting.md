# Scripting — Developer Guide

`ScriptSystem` + `ScriptComponent` let any entity run a [mechanical-basic](https://github.com/mechanical-lich/mechanical-basic) script for custom per-turn and per-interaction behaviour. This replaces bespoke Go systems for entity-specific logic.

> For scripted **status conditions** (on_applied / on_turn / on_death / on_removed), see [Scriptable Conditions](scriptable_conditions.md).

---

## Architecture

| Layer | File |
|-------|------|
| Component definition | `internal/component/ScriptComponent.go` |
| System | `internal/system/ScriptSystem.go` |
| Script files | `data/scripts/*.basic` |
| Blueprint wiring | `ScriptComponent` key in any blueprint JSON |

`ScriptSystem` is registered in `SimWorld.NewSimWorld()`. It runs on every entity that has a `ScriptComponent`.

---

## ScriptComponent

```go
type ScriptComponent struct {
    ScriptPath string         // path to the .basic file, relative to the working directory
    Vars       map[string]any // per-entity variables accessible from the script
}
```

In blueprint JSON:

```json
"ScriptComponent": {
    "ScriptPath": "data/scripts/my_entity.basic",
    "Vars": { "turns_left": 10, "armed": 0 }
}
```

`Vars` is the entity's private state. Scripts read and write it with `get_var` / `set_var`. Values survive across turns. The interpreter is lazily initialised on first use and lives on the component.

---

## Turn lifecycle

`ScriptSystem` integrates with the energy turn system:

1. `AdvanceEnergy` grants `MyTurn` to any entity whose `EnergyComponent.Energy > 0`.
2. `ScriptSystem.UpdateEntity` fires when the entity has `MyTurn`:
   - Adds `TurnTaken` (marks the turn as consumed).
   - Sets `LastActionCost` to `Speed` (so the entity spends exactly what it earns — one action per turn cycle).
   - Calls `on_turn_taken()` in the script, if defined.
3. `CleanUpSystem` calls `ResolveTurn` which sees both `MyTurn` and `TurnTaken`, deducts the energy cost, and removes both markers.

Entities that should tick every turn need `EnergyComponent` and `NeverSleepComponent`:

```json
"EnergyComponent":   { "Speed": 100 },
"NeverSleepComponent": {}
```

Entities that only react to interactions (no per-turn logic) do not need either component.

---

## Script events

### `on_turn_taken()`

Called once per turn when the entity has energy and `MyTurn`. Use this for countdowns, periodic effects, or any time-based logic.

```basic
function on_turn_taken():
    let turns = get_var("turns_left")
    turns = turns - 1
    set_var("turns_left", turns)
    if turns <= 0 then
        add_dead()
    endif
endfunction
```

### `on_interact()`

Called when a player interacts with this entity via the `call_script_interact` interaction trigger. Use this for modal behaviour that needs branching logic beyond what trigger lists can express.

```basic
function on_interact():
    if get_var("armed") = 1 then
        add_message("Already armed.")
        return
    endif
    set_var("armed", 1)
    set_flag("self_destruct_armed", 1)
    add_message("Sequence armed.")
endfunction
```

---

## Built-in functions

### Entity position

| Function | Returns | Description |
|----------|---------|-------------|
| `get_x()` | number | Entity's current X tile coordinate |
| `get_y()` | number | Entity's current Y tile coordinate |
| `get_z()` | number | Entity's current floor (Z) |

### Entity state

| Function | Description |
|----------|-------------|
| `add_dead()` | Marks this entity as dead. `CleanUpSystem` handles removal. |
| `kill_player()` | Adds `DeadComponent` to the player entity. |

### Variables

| Function | Description |
|----------|-------------|
| `get_var(name)` | Reads a value from `ScriptComponent.Vars`. |
| `set_var(name, value)` | Writes a value to `ScriptComponent.Vars`. |

### Level flags

Flags are stored on the `Level` and are readable by win conditions and other systems.

| Function | Description |
|----------|-------------|
| `get_flag(name)` | Reads a value from `Level.Flags`. |
| `set_flag(name, value)` | Writes a value to `Level.Flags`. |

Win conditions check flags by name using a `GameFlag` condition. See [Win Conditions](win_conditions.md).

### Movement

Scripts call these to move the entity one tile per invocation. Each call sets the appropriate `LastActionCost` via the underlying action.

| Function | Returns | Description |
|----------|---------|-------------|
| `move(dx, dy)` | — | Move one step in (dx, dy). Handles bumping, door toggling, and terrain cost. |
| `move_random()` | — | Move in a random cardinal direction. |
| `move_toward(tx, ty)` | 1 or 0 | A* pathfind one step toward (tx, ty). Returns 1 if moved, 0 if stuck/arrived. Supports large (`SizeComponent`) entities. |
| `move_away_from(fx, fy)` | — | Move in the cardinal direction most opposed to (fx, fy). |

### Combat

| Function | Returns | Description |
|----------|---------|-------------|
| `attack(tx, ty)` | — | Melee attack the entity at (tx, ty). |
| `use_skill(ai_type)` | 1 or 0 | Find and execute the skill whose `ai_type` matches the given string. Returns 1 if the skill was found and executed, 0 if the entity has no matching skill. |

### Sensing

| Function | Returns | Description |
|----------|---------|-------------|
| `find_nearest_enemy(range)` | distance or -1 | Search for the nearest non-friendly `BodyComponent` entity within `range` tiles. Returns Euclidean distance, or -1 if none found. Caches the result for `get_target_x/y`. |
| `get_target_x()` | number | X coordinate of the last `find_nearest_enemy` result. |
| `get_target_y()` | number | Y coordinate of the last `find_nearest_enemy` result. |
| `distance_to(tx, ty)` | number | Euclidean distance from entity to (tx, ty). |
| `has_los(tx, ty)` | 1 or 0 | 1 if there is clear line of sight to (tx, ty) on the current floor. |
| `is_aligned(tx, ty)` | 1 or 0 | 1 if the entity shares a row or column with (tx, ty) and has clear LOS (ranged alignment check). |
| `friendly_on_ray(tx, ty)` | 1 or 0 | 1 if a friendly entity lies on the straight line between the entity and (tx, ty). Used for friendly-fire avoidance. |
| `get_faction()` | string | The entity's `DescriptionComponent.Faction` (e.g. `"plantz"`, `"crew"`). |

### World interaction

| Function | Description |
|----------|-------------|
| `spawn_entity(blueprint, x, y, z)` | Creates and places a new entity. |
| `add_message(text)` | Posts a message to the player's message log. |
| `num_to_str(n)` | Converts a number to a string for concatenation. |
| `rnd_int(max)` | Returns a random integer in `[0, max)`. |

---

## AI template scripts

Entity AI is implemented as shared template scripts in `data/scripts/`. Blueprints supply configuration through `ScriptComponent.Vars` rather than bespoke Go systems.

### `ai_wander.basic`

Moves randomly every turn. No configuration needed.

### `ai_hostile.basic`

Hunts the nearest enemy within `sight_range`. Attacks when adjacent; prefers `align_and_shoot` and `melee_skill` skills when available.

| Var | Type | Description |
|-----|------|-------------|
| `sight_range` | number | Detection radius in tiles. |

### `ai_advanced.basic`

Full idle / hunt / flee state machine with action scoring.

| Var | Type | Description |
|-----|------|-------------|
| `aggressiveness` | 0–100 | Chance per tick to enter `hunt` when a target is visible. Proximity bonus adds up to +40. |
| `fear` | 0–100 | Chance per tick to enter `flee`. Overrides `aggressiveness` when higher. |
| `sight_range` | number | Detection radius in tiles. |
| `randomness` | number | Jitter added to action scores. Higher = more varied choices. |
| `avoids_friendly_fire` | 0 or 1 | Skip `align_and_shoot` if a friendly is on the ray. |

In idle state the script tries `use_skill("spread_overgrowth")` before wandering, so spreader-type entities work without any extra configuration.

**Blueprint example:**

```json
"ScriptComponent": {
    "ScriptPath": "data/scripts/ai_advanced.basic",
    "Vars": {
        "aggressiveness": 60,
        "fear": 20,
        "sight_range": 12,
        "randomness": 15,
        "avoids_friendly_fire": 1
    }
}
```

### Comment syntax

mechanical-basic uses `#` for comments, not `'`. A script with `'` comments will fail to parse and the entity will loop indefinitely (it holds `MyTurn` but never gains `TurnTaken`).

```basic
# This is a valid comment
' This will cause a parse error
```

---

## Wiring `on_interact`

To call `on_interact` from an interaction, add `call_script_interact` to the entity's `InteractionComponent` triggers:

```json
"InteractionComponent": {
    "Prompt": "Arm the sequence?",
    "Repeatable": false,
    "Triggers": [
        { "Type": "call_script_interact", "Params": {} }
    ]
}
```

The listener (`internal/game/listeners/interaction.go`) calls `system.CallScriptEvent("on_interact", entity, level)`.

---

## Existing scripts

| File | Used by | Description |
|------|---------|-------------|
| `data/scripts/ai_wander.basic` | — | Random cardinal move each turn. |
| `data/scripts/ai_hostile.basic` | `abomination`, `viner`, `scythe`, `scrambler`, `spitter` | Hunts nearest enemy; attacks or uses skill when adjacent. |
| `data/scripts/ai_advanced.basic` | `creeper`, `spreader`, `ingrained_spreader`, `massive_spreader`, `mother_plant`, `crewmember`, `officer` | Full idle/hunt/flee state machine with action scoring. |
| `data/scripts/mobile_mother_plant.basic` | `mobile_mother_plant` | Decrements `turns_left` each turn; at zero, kills self and spawns `mother_plant`. |
| `data/scripts/self_destruct_console.basic` | `self_destruct_console` | `on_interact` arms the countdown; `on_turn_taken` decrements it, warns at ≤10, calls `kill_player()` at zero. |

---

## Adding a new scripted entity

**For a new enemy or NPC** — use a template script and supply config through `Vars`:

1. Add `ScriptComponent` pointing to the appropriate template (`ai_hostile.basic`, `ai_advanced.basic`, or `ai_wander.basic`) with the required `Vars`.
2. Add `EnergyComponent` (Speed) so the entity takes turns.
3. No new Go code needed.

**For a unique interactive object** (console, machine, event trigger):

1. Create `data/scripts/<name>.basic` and implement `on_turn_taken` and/or `on_interact`.
2. Add `ScriptComponent` to the blueprint with the script path and initial `Vars`.
3. If the entity needs per-turn logic, add `EnergyComponent` (Speed) and `NeverSleepComponent`.
4. If the entity needs an interaction dialog, add `InteractionComponent` with `call_script_interact` trigger.
