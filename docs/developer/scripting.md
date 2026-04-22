# Scripting — Developer Guide

`ScriptSystem` + `ScriptComponent` let any entity run a [mechanical-basic](https://github.com/mechanical-lich/mechanical-basic) script for custom per-turn and per-interaction behaviour. This replaces bespoke Go systems for entity-specific logic.

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

### World interaction

| Function | Description |
|----------|-------------|
| `spawn_entity(blueprint, x, y, z)` | Creates and places a new entity. |
| `add_message(text)` | Posts a message to the player's message log. |
| `num_to_str(n)` | Converts a number to a string for concatenation. |
| `rnd_int(max)` | Returns a random integer in `[0, max)`. |

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

| File | Entity | Description |
|------|--------|-------------|
| `data/scripts/mobile_mother_plant.basic` | `mobile_mother_plant` | Decrements `turns_left` each turn; at zero, kills self and spawns `mother_plant`. |
| `data/scripts/self_destruct_console.basic` | `self_destruct_console` | `on_interact` arms the countdown; `on_turn_taken` decrements it, warns at ≤10, calls `kill_player()` at zero. |

---

## Adding a new scripted entity

1. Create `data/scripts/<name>.basic` and implement `on_turn_taken` and/or `on_interact`.
2. Add `ScriptComponent` to the blueprint JSON with the script path and initial `Vars`.
3. If the entity needs per-turn logic, also add `EnergyComponent` (Speed) and `NeverSleepComponent`.
4. If the entity needs an interaction dialog, add `InteractionComponent` with `call_script_interact` trigger.
