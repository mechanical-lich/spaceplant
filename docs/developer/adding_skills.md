# Adding Skills and Actions

## Overview

Skills are defined in `data/skills/skills.json` and loaded at startup. A skill can do any combination of:

- Modify the entity's stats passively (e.g. +2 AC, fire resistance)
- Bind a key to an action when the skill is active
- Tell the AI how to use the skill autonomously

---

## Step 1: Add an entry to `data/skills/skills.json`

The file is a JSON array of skill objects. Add a new entry:

```json
{
  "id": "my_skill",
  "name": "My Skill",
  "description": "Does something cool. Bound to X.",
  "action_bindings": {
    "x": "some_action"
  },
  "action_params": {
    "damage_dice": "2d6",
    "damage_type": "fire",
    "cool_dc": 14
  },
  "ai_type": "align_and_shoot",
  "stat_modifiers": [
    { "stat": "cs", "delta": 5 }
  ]
}
```

All fields except `id`, `name`, and `description` are optional.

---

## Skill Fields

| Field | Type | Description |
|---|---|---|
| `id` | string | Unique identifier. Used in blueprints, character creator config, and code lookups. |
| `name` | string | Display name shown in UI. |
| `description` | string | Shown in the character creator skill tab. |
| `action_bindings` | map | Keys the player presses to trigger the action. Format: `{ "key": "action_id" }`. |
| `action_params` | object | Data passed to the action at runtime (see below). |
| `ai_type` | string | How the AI uses this skill. See [AI Types](#ai-types). |
| `stat_modifiers` | array | Passive stat changes applied when the skill is active. See [Stat Modifiers](#stat-modifiers). |

---

## Action Params

`action_params` are passed to the action when it executes, allowing a single action (like `cone_of`) to behave differently depending on the skill that triggers it.

| Field | Type | Description |
|---|---|---|
| `damage_dice` | string | Damage roll expression, e.g. `"2d6"`, `"1d8"`. |
| `damage_type` | string | Type of damage: `"fire"`, `"acid"`, `"poison"`, `"slashing"`, etc. |
| `cool_dc` | int | Difficulty modifier for the Cool Check resistance roll. Higher = harder to resist. See [combat.md](combat.md) for the formula. |
| `depth` | int | Depth in tiles for cone effects. |
| `spread` | int | Cone width. `0` = line only. `-1` = classic widening cone (spreads by 1 per depth row). |
| `range` | int | Max reach in tiles for targeted or line effects. |
| `radius` | int | Radius for circular area effects. |
| `verb` | string | Word used in combat messages, e.g. `"bite"`, `"stab"`. |
| `extra_damage_on_failed_save` | string | Bonus damage dice when the target fails their Cool Check, e.g. `"3d6"`. |
| `status_condition_on_fail_save` | string | Status applied on failed Cool Check. `"poison"` and `"burning"` use `DamageConditionComponent` (damage per turn from `extra_damage_on_failed_save`, type matching the condition name). `"slowed"` and `"haste"` use their typed speed-modifier components. |
| `status_condition_duration` | int | How many turns the status condition lasts. |
| `action_cost` | int | Overrides the default energy cost. `0` (unset) uses the action's default (`CostAttack` = 100). Set to `50` for a half-action (`CostQuick`). |

---

## Stat Modifiers

Stat modifiers are applied when the skill is granted and reversed when it is removed.

```json
"stat_modifiers": [
  { "stat": "ac", "delta": 2 },
  { "stat": "resistance", "value": "fire" },
  { "stat": "advantage", "value": "dex" }
]
```

**Numeric stats** — use `delta`:

| Stat | Effect |
|---|---|
| `speed` | Action frequency. Base is 100. +50 acts 50% more often. |
| `ph` | Physique. Increases bare-hands Penetration. |
| `ag` | Agility. |
| `ma` | Mental Ability. |
| `cl` | Cool. Improves Cool Check resistance rolls. |
| `ld` | Leadership. |
| `cs` | CombatSkill. Directly raises percentile hit chance. |
| `htcs` | Hand-to-Hand Combat Skill. Raises melee to-hit and parry chance. |
| `natural_sp` | Natural Stopping Power. Absorbs incoming damage regardless of armor worn. |

**String-set stats** — use `value`:

| Stat | Effect |
|---|---|
| `resistance` | Half damage from the given damage type. |

---

## AI Types

`ai_type` controls how the AI system uses the skill. Without it the AI ignores the skill.

| Value | Behaviour |
|---|---|
| `align_and_shoot` | AI gets in line-of-sight range, aligns to the same row or column as the target, then fires. Used for cone and line attacks. |
| `melee_skill` | AI uses the skill instead of a basic attack when adjacent to the target. |

---

## Step 2: Grant the skill to an entity

**Via blueprint** (`data/blueprints/*.json`):
```json
"SkillComponent": {
  "Skills": ["my_skill"]
}
```

**Via character creator config** (`data/character_creator_config.json`) — skills listed under `base_skills` appear as always-available options in the character creator:
```json
{
  "base_skills": ["thick_skin", "brawler", "my_skill"]
}
```

**Via class definition** — add the skill ID to a class's `StartingSkills` (granted automatically) or `Skills` (available to purchase) list in `data/classes/classes.json`.

---

## Adding a Class

Classes are defined in `data/classes/classes.json`. Each entry supports:

| Field | Type | Description |
|---|---|---|
| `ID` | string | Unique identifier used in code and saves. |
| `Name` | string | Display name. |
| `Description` | string | Shown in the character creator. |
| `StatMods` | array | Direct stat bonuses applied at spawn, independent of skills. Use `{ "stat": "ma", "delta": 5 }`. Same stat keys as skill stat modifiers. |
| `StartingItems` | array | Blueprint IDs to spawn and equip when the game starts. Processed in order; `EquipAllBest` runs after all items are added. |
| `StartingSkills` | array | Skill IDs granted automatically when the class is assigned. |
| `Skills` | array | Skill IDs the player can purchase via upgrade points. |

Example:

```json
{
    "ID": "scientist",
    "Name": "Scientist",
    "Description": "A researcher trained in advanced sciences.",
    "StatMods": [
        { "stat": "ma", "delta": 5 }
    ],
    "StartingItems": [
        "laser_trimmers",
        "lab_coat",
        "scrubs"
    ],
    "StartingSkills": [],
    "Skills": ["brainy", "nimble", "thick_skin"]
}
```

**StatMods vs skill stat modifiers:** `StatMods` on a class are applied once at spawn directly to the `StatsComponent`. They are not reversible at runtime and do not require a skill entry. Use them for the class's core identity bonuses. Use skill `stat_modifiers` when the bonus needs to be tied to a specific skill (so it can be removed if the skill is taken away).

**Minimap reveal:** If a class grants the `completed_minimap` skill via `StartingSkills`, `SpawnPlayer` will automatically call `RevealFloor(0)` to pre-populate the minimap.

---

## Available Actions

These are the registered action IDs you can reference in `action_bindings`:

| Action ID | Description | Uses ActionParams |
|---|---|---|
| `cone_of` | Fires a cone of the given damage type in the direction the entity is facing. All entities in the cone make a saving throw; damage on fail. | Yes — `depth`, `spread`, `damage_dice`, `damage_type`, `save_stat`, `save_dc` |
| `roundhouse_kick` | Attacks all 8 adjacent tiles for standard melee damage. | No |
| `shove` | Pushes the entity directly in front one tile further. Cannot shove massive or inanimate entities. | No |
| `bite` | Attacks the entity directly in front for 1d8 damage. On hit, target makes a CON save (DC 13) or becomes poisoned for 5 turns. | No |
| `heal` | Uses the first healing item in the entity's inventory. | No |
| `pickup` | Picks up the first item at the entity's current tile. | No |
| `equip` | Auto-equips the first equippable item in the entity's inventory. | No |
| `stairs` | Uses stairs at the entity's current tile. | No |
