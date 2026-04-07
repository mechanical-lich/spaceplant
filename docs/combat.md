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
| Hand-to-Hand Combat Skill | HTCS | 1–100 | Percentile hit chance used for melee attacks and parry rolls. |

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

## Melee Combat

Melee attacks use a separate percentile stat, `HTCS` (Hand-to-Hand Combat Skill). The effective melee to-hit is calculated as:

```
EffectiveMeleeCS = HTCS + PHBracketBonus
```

`PHBracketBonus` is a small bracketed bonus derived from the attacker's `PH` (Physique). In code this is implemented by `statMeleeBonus()` and follows Phoenix Command-style brackets (for example: PH ≥ 18 => +20, PH ≥ 16 => +15, PH ≥ 14 => +10, PH ≥ 12 => +5, PH ≥ 10 => +0, otherwise −10).

Parry: after damage is calculated (Penetration minus Stopping Power), the defender gets a parry roll if the attack is melee. The defender's parry threshold is:

```
ParryThreshold = Defender.HTCS + AGBracketBonus
```

Where `AGBracketBonus` is the same bracket function described above applied to `AG` (Agility). The game rolls `1d100`; if the roll is less than or equal to `ParryThreshold` then the attack is considered parried:
- The game halves the final damage (integer division) and marks the combat event with `Parried = true`.
- If the final damage was already 0 (armor fully absorbed Pen), the result is a full parry (attack landed but did no damage).

In short:
- Melee offense uses `HTCS + PH bracket bonus`.
- Melee defense/parry uses `HTCS + AG bracket bonus` and may halve or negate damage.

Combat events include a `Parried` flag to indicate this outcome; UI messaging shows distinct lines for full parries (no damage) and partial blocks (damage reduced).

See code locations for the exact implementation: `internal/combat/combat.go` (melee CS and parry), `internal/component/stats.go` (new `HTCS` field), and `internal/combat/event.go` (`Parried` flag).

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
| Security pistol | 15 |
| Assault rifle | 14 |
| Shotgun | 18 |

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

If an entity has both a resistance and a weakness to the same type, the effects cancel out and damage is unchanged.

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
| `AttackRange` | int | Maximum range in tiles for melee weapons (0 or 1 = adjacent only). |
| `Range` | int | Maximum range in tiles for ranged weapons. |
| `Ranged` | bool | If true, weapon can only be used via the shoot action (F / Shift+F / G). |
| `BurstSize` | int | Rounds fired per burst (G key). 0 or 1 means single-shot only; ≥2 enables burst fire. |
| `SpreadAngle` | int | Extra parallel fire lines on each snap/aimed shot. 0 = single line; 1 = 3-wide; 2 = 5-wide. Spread lines deal 60% Pen. |

---

## Ranged Combat

When a ranged weapon is equipped, pressing **F** fires a snap shot, **Shift+F** fires an aimed shot, and **G** fires a burst (if the weapon supports it). All shots travel in the direction the player is currently facing. Facing can be changed with **Shift+W/A/S/D** without spending an action.

The shot traces a line tile-by-tile along the facing direction and hits the first solid non-self entity within range. Solid tiles (walls, doors) stop the shot. A brief visual tracer is drawn along the bullet path.

### Range Bands

The attacker's effective CS is modified by the distance to the target:

| Band | Distance | CS modifier |
|---|---|---|
| Point blank | ≤ 1 tile | −20 |
| Effective range | 2 to Range/2 tiles | ±0 |
| Long range | > Range/2 tiles | −15 |

### Snap Shot vs Aimed Shot vs Burst

| Mode | Key | Energy cost | Notes |
|---|---|---|---|
| Snap shot | F | 100 | Single shot. SpreadAngle applies. |
| Aimed shot | Shift+F | 150 | +10 CS. SpreadAngle applies. |
| Targeted aimed shot | Shift+F | 200 | +10 CS. Opens a modal to choose which body part to aim at. |
| Burst fire | G | 150 | Fires `BurstSize` rounds along the same line. Requires `BurstSize ≥ 2`. |

**Burst fire** rolls independently for each round. The first round gets +15 CS (burst bonus); subsequent rounds lose 5 CS each from recoil (+10, +5, …). If the first round misses into the void, the burst stops early.

#### Targeted Aimed Shot

When `Shift+F` is pressed and a valid target is in the line of fire, a modal lists the target's non-amputated body parts. The player selects one (by clicking or pressing the number hotkey shown). The shot then resolves normally but with a CS-scaled chance of striking the chosen part:

```
HitChance(chosen part) = 75 + (CS − 50) / 10    [clamped to 60–90%]
```

If the bias roll fails, the hit location falls back to normal weighted random selection. If the chosen part is amputated before the shot resolves, the bias is skipped entirely. Spread-angle pellets (secondary lines) always use random hit location.

If nothing is in the line of fire, a "Nothing to aim at." message is logged and no action is taken.

The full CS for a single ranged attack is:

```
EffectiveCS = CS + WeaponCombatSkillModifier + RangeBandModifier + ShotModeBonus
```

**Shot mode bonuses:**

| Mode | CS bonus |
|---|---|
| Snap shot | 0 |
| Aimed shot | +10 |
| Burst round 1 | +15 |
| Burst round 2 | +10 |
| Burst round 3 | +5 |

### Facing and Direction

Moving in any direction automatically updates facing. Pressing **Shift+direction** rotates the player without moving or consuming a turn, which is useful for pre-aiming before firing.

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
