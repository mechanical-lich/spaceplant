# Skill Chips — Developer Guide

Skill chips are consumable items that permanently teach the player a skill when used from the inventory. There is a single generic `skill_chip` blueprint; the target skill is configured at spawn time by the setup script.

---

## Architecture

| Layer | File |
|---|---|
| Component | `internal/component/SkillChipComponent.go` |
| Component type constant | `internal/component/component_types.go` (`SkillChip`) |
| Factory registration | `internal/factory/component_registry.go` |
| Consume logic | `internal/game/characterStatsView.go` (`onInvEquip`) |
| Inventory description | `internal/game/characterStatsView.go` (`refreshInvDesc`) |
| Blueprint | `data/blueprints/items/misc/skill_chips.json` (`skill_chip`) |
| Script function | `internal/system/setup_script.go` (`spawn_skill_chip`) |
| Spawn script | `data/scripts/scenarios/skill_chips_setup.basic` |

---

## How It Works

1. `spawn_skill_chip(skillId, z, room)` in a setup script creates the generic `skill_chip` entity, sets `SkillChipComponent.SkillId`, and patches the item's display name and description to reflect the actual skill. The skill name is resolved from `skill.Get(skillId)`.
2. When the player selects the chip in **Inventory → Equip/Use**, `onInvEquip` checks `Effect == "skill_chip"`.
3. If the player already has the skill, a message is logged and nothing is consumed.
4. Otherwise, `skill.Apply(player, skillId)` grants the skill (applying all stat modifiers and action bindings), the chip is removed from inventory, and a message is logged.

---

## Spawning Chips

Use `spawn_skill_chip(skill_id, z, room_index)` in any setup script. The skill ID must match an entry in `data/skills/skills.json`.

```
function on_setup():
    let f = random_floor_excluding(-1)
    spawn_skill_chip("brawler", f, random_room(f))

    let f = random_floor_excluding(-1)
    spawn_skill_chip("sharpshooter", f, random_room(f))
endfunction
```

`random_floor_excluding(-1)` picks any floor (since -1 matches no real floor index).

The plantz scenario runs `data/scripts/scenarios/skill_chips_setup.basic` which currently scatters 12 chips across the station. To add more chips or change which skills appear, edit that script. To wire a script into a scenario, add it to `setup_scripts` in the scenario JSON:

```json
"setup_scripts": [
    "data/scripts/scenarios/self_destruct_setup.basic",
    "data/scripts/scenarios/skill_chips_setup.basic"
]
```

---

## Scripting Reference

### `spawn_skill_chip(skill_id, z, room_index)`

Creates a skill chip for the given skill and places it at a random floor tile in the specified room.

| Argument | Type | Description |
|---|---|---|
| `skill_id` | string | ID of the skill to teach. Must exist in `data/skills/skills.json`. |
| `z` | int | Floor index (0-based). |
| `room_index` | int | Room index on that floor, as returned by `random_room` or `find_room_by_tag`. |

Errors if the blueprint `skill_chip` cannot be created, the floor/room is out of range, or no empty tile is found after 80 attempts.
