# Combat System

Spaceplant uses a percentile-based combat system inspired by the Aliens Adventure Game (Leading Edge Games, 1991), itself derived from the Phoenix Command rulebook. The core philosophy is simulation over abstraction: attacks land based on skill, damage goes to specific body parts, and armor absorbs rather than prevents damage outright.

---

## Character Stats

Every entity has a `StatsComponent` with six attributes:

| Stat | Abbrev | Range | Role |
|---|---|---|---|
| Physique | PH | 3–18 | Physical power. Determines bare-hands Penetration bonus. |
| Agility | AG | 3–18 | Speed and coordination. |
| Mental Ability | MA | 3–18 | Intelligence. |
| Cool | CL | 3–18 | Composure under fire. Drives Cool Check resistance rolls. |
| Leadership | LD | 3–18 | Command presence. |
| CombatSkill | CS | 1–100 | Percentile hit chance. The primary combat stat. |

Typical values for reference:

| Entity type | CS |
|---|---|
| Untrained crew | 25–35 |
| Average crewmember | 40 |
| Officer | 40 |
| Player (default) | 40 |
| Experienced monster | 45–50 |
| Elite combatant | 60–70 |

---

## To-Hit Resolution

Each attack rolls `d100`. The attack **hits** if:

```
d100 <= CombatSkill + WeaponCombatSkillModifier
```

- `CombatSkill` comes from the attacker's `StatsComponent.CS`.
- `WeaponCombatSkillModifier` comes from the equipped weapon's `CombatSkillModifier` field (can be positive or negative). It is 0 for bare-hands attacks.
- There are no crits or fumbles — the roll is a straight percentile check.

---

## Damage: Penetration vs Stopping Power

When a hit lands, damage is resolved as:

```
Damage = max(0, Penetration - StoppingPower)
```

**Penetration (Pen)** is the weapon's raw damage value. It does not vary per hit — it is a fixed property of the weapon or attack type.

| Attack type | Pen |
|---|---|
| Bare hands (PH 10) | ~5–6 |
| Bare hands (PH 18) | ~8 |
| Laser trimmers | 10 |

Bare-hands Pen is calculated as `3 + PH / 4`.

**Stopping Power (SP)** is the armor's damage absorption. It comes from two sources that are added together:
- `NaturalSP` on the hit entity's `StatsComponent` (from skills like *Brawler* or *Thick Skin*).
- The `StoppingPower` of any armor item equipped on the body part that was hit.

| Armor tier | SP |
|---|---|
| No armor (scrubs, lab coat, etc.) | 0 |
| Light protection (engineer shirt, work trousers, officer cap) | 3 |
| Heavy protection (security gear, helmet, hardhat, workboots) | 6 |

Damage of 0 is possible — the attack landed but the armor fully absorbed it.

---

## Hit Location and Body Parts

Every entity has a `BodyComponent` with named body parts, each with its own HP pool and flags. When a hit lands, the game selects which body part is struck.

**Selection** uses the entity's `HitLocationComponent`, which maps part names to integer weights. A part with weight 40 is four times as likely to be hit as one with weight 10. If the `HitLocationComponent` is absent, all living parts are equally likely.

Damage is applied to the chosen part's HP. When a part's HP reaches 0:

- `Broken = true` is set on the part.
- If `KillsWhenBroken` is true, the entity dies.
- If `KillsWhenAmputated` is true and the part HP went negative, it is also amputated and the entity dies.

The head and torso of most entities kill on break. Limbs generally do not.

**Example — human body layout:**

| Part | Weight | HP | Kills on break |
|---|---|---|---|
| torso | 40 | varies | yes |
| right_arm | 15 | varies | no |
| left_arm | 15 | varies | no |
| legs | 15 | varies | no |
| head | 10 | varies | yes |
| feet | 5 | varies | no |

---

## Cool Checks

Cool Checks replace saving throws for any skill or effect that an entity might resist. The formula is:

```
success if d100 <= (CL * 5) - cool_dc
```

- `CL` is the target entity's Cool stat. At CL 10 (average), the base threshold is 50%.
- `cool_dc` is a difficulty modifier set by the attack or skill. Higher values make the check harder.
- If the threshold drops to 0 or below, the check still has a 1% chance of success.

Cool Checks are used by:
- Cone attacks (flamethrower, acid spray) — target resists full damage on success.
- Melee specials (poisonous bite, paralyzing vinewhip) — target resists the status effect on success.

---

## Damage Types, Resistances, and Weaknesses

Every attack has a damage type string. Current types in use: `slashing`, `bludgeoning`, `fire`, `acid`, `poison`.

An entity can have:
- **Resistances** — takes half damage from listed types.
- **Weaknesses** — takes double damage from listed types.

These are stored as `[]string` on `StatsComponent` and are also granted by skills (see *Noncombustible*, *Acid Proof*).

If an entity has both a resistance and a weakness to the same type, the resistance wins.

---

## Equipped Armor and Weapons

Entities use `BodyInventoryComponent` from ml-rogue-lib to manage worn items. Items are keyed to body part slots — only the armor item worn on the part that was hit contributes its SP to that attack.

Weapons are equipped to hand slots. The combat system checks `BodyInventoryComponent.Equipped` first, then falls back to `InventoryComponent.RightHand` / `LeftHand`.

**Weapon fields relevant to combat:**

| Field | Type | Description |
|---|---|---|
| `Penetration` | int | Pen value. Base damage if armor is zero. |
| `CombatSkillModifier` | int | Bonus or penalty to the wielder's CS for this attack. |
| `DamageType` | string | Damage type string passed to resistance checks. |
| `AttackRange` | int | Maximum range in tiles (0 or 1 = melee only). |

**Armor fields relevant to combat:**

| Field | Type | Description |
|---|---|---|
| `StoppingPower` | int | SP value. Absorbed from incoming Pen before damage is applied. |
| `Resistances` | []string | Additional damage type resistances granted while worn. |

---

## Skills That Affect Combat

### Passive stat skills

| Skill | Effect |
|---|---|
| Brawler | +2 NaturalSP, +1 PH |
| Thick Skin | +2 NaturalSP |
| Noncombustible | Resistance to fire |
| Acid Proof | Resistance to acid |
| Sharpshooter | +5 CS |
| Combat Specialist | +5 CS |
| Hands Only | Unarmed attacks use a higher Pen value and cost half energy |

### Active skills with status conditions

Skills that use `melee_special` or `cone_of` actions can apply status conditions when the target fails a Cool Check.

| Skill | Action | On failed Cool Check |
|---|---|---|
| Poisonous Bite | Melee | Target becomes **poisoned** for 5 turns (damage per turn, type: poison) |
| Paralyzing Vinewhip | Melee | Target becomes **slowed** for 4 turns (reduced action frequency) |
| Flamethrower | Cone (depth 3, widening) | Target takes full fire damage |
| Acid Spray | Cone (depth 5, line) | Target takes full acid damage |

Status conditions:

| Condition | Mechanic |
|---|---|
| `poison` | Applied via `DamageConditionComponent`. Deals damage each turn for the specified duration using the `damage_dice` and `damage_type` from the skill's `action_params`. |
| `burning` | Same as poison but with fire damage type. |
| `slowed` | Applied via `SlowedComponent`. Reduces the entity's action speed for the duration. |
| `haste` | Applied via `HasteComponent`. Increases action speed for the duration. |

The duration of all conditions is set by `status_condition_duration` in the skill definition.
