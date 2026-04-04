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
    "save_stat": "dex",
    "save_dc": 14
  },
  "ai_type": "align_and_shoot",
  "stat_modifiers": [
    { "stat": "ac", "delta": 1 }
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
| `save_stat` | string | Ability score for saving throw: `"str"`, `"dex"`, `"con"`, `"int"`, `"wis"`. |
| `save_dc` | int | Difficulty class for the saving throw. |
| `depth` | int | Depth in tiles for cone effects. |
| `spread` | int | Cone width. `0` = line only. `-1` = classic widening cone (spreads by 1 per depth row). |
| `range` | int | Max reach in tiles for targeted or line effects. |
| `radius` | int | Radius for circular area effects. |
| `verb` | string | Word used in combat messages, e.g. `"bite"`, `"stab"`. |
| `extra_damage_on_failed_save` | string | Bonus damage dice when the target fails their save, e.g. `"3d6"`. |
| `status_condition_on_fail_save` | string | Status applied on failed save. `"poison"` and `"burning"` use `DamageConditionComponent` (damage per turn from `extra_damage_on_failed_save`, type matching the condition name). `"slowed"` and `"haste"` use their typed speed-modifier components. |
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
| `ac` | Armor class. |
| `str` | Strength. |
| `dex` | Dexterity. |
| `con` | Constitution. |
| `int` | Intelligence. |
| `wis` | Wisdom. |
| `melee_attack_bonus` | Bonus added to melee to-hit rolls. |
| `ranged_attack_bonus` | Bonus added to ranged to-hit rolls. |

**String-set stats** — use `value`:

| Stat | Effect |
|---|---|
| `resistance` | Half damage from the given damage type. |
| `advantage` | Roll twice and take the best on saves for the given ability score. |

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

**Via class definition** — add the skill ID to the class's skill list in `data/classes/`.

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
