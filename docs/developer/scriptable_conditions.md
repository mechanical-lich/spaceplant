# Scriptable Conditions — Developer Guide

`ScriptableConditionComponent` is a decaying status condition driven by a mechanical-basic script. It slots into the existing `ActiveConditionsComponent` system alongside `DamageConditionComponent` and `StatConditionComponent`, but lets you define arbitrary per-turn and death behaviour in data rather than Go.

---

## When to use it

| Use case | Approach |
|----------|----------|
| Fixed periodic damage (poison, burning) | `DamageConditionComponent` — simpler |
| Stat buff/debuff while active | `StatConditionComponent` — simpler |
| Custom logic on turn, death, or expiry | **`ScriptableConditionComponent`** |

Good fits: conditions that spawn entities, check health thresholds, play messages based on state, chain into other conditions, or react to death.

---

## Script events

A condition script is a `.basic` file that may define any of these functions. All are optional.

| Function | When it fires |
|----------|---------------|
| `on_applied()` | Once, when the condition is first added to the entity |
| `on_turn()` | Each turn the entity acts (respects `interval`) |
| `on_death()` | When the host entity dies while the condition is active |
| `on_removed()` | When the condition expires naturally (duration reaches 0) |

```basic
function on_applied():
    add_message("Roots take hold...")
endfunction

function on_turn():
    deal_damage(1, "poison")
endfunction

function on_death():
    spawn_entity("scrambler", get_x(), get_y(), get_z())
    add_message("The roots convulse — something emerges!")
endfunction

function on_removed():
    add_message("The roots wither away.")
endfunction
```

---

## Built-in functions

### Position

| Function | Returns | Description |
|----------|---------|-------------|
| `get_x()` | number | Entity's X tile coordinate |
| `get_y()` | number | Entity's Y tile coordinate |
| `get_z()` | number | Entity's floor (Z) |

### Variables

Condition-local state. Values persist across turns for the lifetime of the condition.

| Function | Description |
|----------|-------------|
| `get_var(key)` | Read a value from the condition's `Vars` map |
| `set_var(key, value)` | Write a value to the condition's `Vars` map |
| `get_duration()` | Remaining turns before the condition expires |

### Messages

| Function | Description |
|----------|-------------|
| `add_message(text)` | Post a message to the player's message log |

### Damage

| Function | Description |
|----------|-------------|
| `deal_damage(amount, type)` | Deal `amount` damage to the host entity (hits a random body part, or HP if no body) |
| `apply_damage_condition(name, duration, dice, type)` | Add a `DamageConditionComponent` to the entity — useful for chaining into standard conditions |

### World

| Function | Description |
|----------|-------------|
| `spawn_entity(blueprint, x, y, z)` | Create and place a new entity at the given position. Requires the condition to have been ticked (not called from `on_applied`). |

---

## Applying via an action

Add `status_condition_script` and `status_condition_duration` to a skill's `action_params`. The condition is applied to the target when they fail their save (`resist_dc` > 0).

```json
{
  "id": "rooting_strike",
  "name": "Rooting Strike",
  "ai_type": "melee_skill",
  "action_bindings": { "b": "melee_special" },
  "action_params": {
    "damage_dice": "1d6",
    "damage_type": "physical",
    "verb": "strike",
    "resist_dc": 12,
    "check_stat": "ph",
    "status_condition_on_fail_save": "rooted",
    "status_condition_duration": 8,
    "status_condition_script": "data/scripts/conditions/rooted.basic"
  }
}
```

If the script path does not exist, a warning is printed to the log and no condition is applied. Supported actions: `melee_special`, `cone_of`.

### `action_params` reference

| Field | Type | Description |
|-------|------|-------------|
| `status_condition_script` | string | Path to the `.basic` script, relative to the working directory |
| `status_condition_duration` | int | How many turns the condition lasts |
| `status_condition_script_interval` | int | Fire `on_turn` every N turns. Default: 1 (every turn) |
| `status_condition_on_fail_save` | string | Display name for the condition (used for deduplication) |

---

## Applying via blueprint

You can attach a `ScriptableConditionComponent` directly to an entity in its blueprint JSON. This is useful for permanent or starting conditions.

```json
"ScriptableConditionComponent": {
    "Name": "cursed",
    "Duration": 20,
    "ScriptPath": "data/scripts/conditions/curse.basic",
    "Interval": 3,
    "Vars": { "damage_amount": 2 }
}
```

Note: a blueprint `ScriptableConditionComponent` sits directly on the entity, not inside `ActiveConditionsComponent`. It will still receive `on_applied`, `on_turn`, `on_death`, and `on_removed` via the same system.

---

## Passing configuration into scripts

Use `Vars` to make a script reusable across different conditions without duplicating the `.basic` file.

```json
"status_condition_script": "data/scripts/conditions/periodic_damage.basic",
"status_condition_duration": 6
```

`periodic_damage.basic` reads `damage_amount` from `Vars` and defaults to 1 if not set:

```basic
function on_turn():
    let amt = get_var("damage_amount")
    if amt == 0 then amt = 1
    let dtype = get_var("damage_type")
    if dtype == "" then dtype = "physical"
    deal_damage(amt, dtype)
endfunction
```

---

## Death hook timing

`on_death` fires in `CleanUpSystem` on the **first frame** an entity receives `DeadComponent`, before `Solid`, `Energy`, and `MyTurn` are stripped. This means `spawn_entity` works correctly — the replacement entity enters the level immediately and starts taking turns next cycle.

If an entity dies with multiple scriptable conditions, each `on_death` fires in the order the conditions were added.

---

## Existing condition scripts

| File | Used by | Description |
|------|---------|-------------|
| `data/scripts/conditions/periodic_damage.basic` | generic | Deals `damage_amount` damage per turn. Configure via `Vars`. |
| `data/scripts/conditions/rooted.basic` | `rooting_strike` skill | No damage. On death, spawns a scrambler at the entity's position. |
