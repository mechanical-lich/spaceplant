# Combat

## How Attacks Work

Attacks are resolved with a percentile roll against your **Combat Skill (CS)** for ranged attacks, or **Hand-to-Hand Combat Skill (HTCS)** for melee. If the roll is equal to or under your skill, the attack hits.

Weapons have their own accuracy modifiers, and distance affects ranged shots.

## Ranged Combat

Fire in the direction you are facing. You must have a compatible weapon equipped and, if it uses a magazine, ammo loaded.

| Mode | Key | Energy | Accuracy | Notes |
|---|---|---|---|---|
| Snap shot | F | Low | Standard | Quick shot. Spread weapons fire a wide pattern. |
| Aimed shot | Shift+F | Medium | +10 CS | More accurate. Costs more energy. |
| Targeted aimed shot | Shift+F | High | +10 CS | Opens a menu to choose which body part to aim for. |
| Burst fire | G | Medium | +15 CS (first round) | Fires multiple rounds rapidly. Accuracy drops after the first shot. |

> **Targeted aimed shot:** When you press Shift+F and there is an enemy in your line of fire, a menu appears listing available body parts. Select one by clicking or pressing the number key shown (1–9). Your shot has a high chance of hitting the chosen part — modified by your CS score. If nothing is in front of you, a message is logged instead.

### Distance Modifiers

Accuracy changes based on how far away the target is:

| Distance | CS Modifier |
|---|---|
| Point blank (≤ 1 tile) | −20 |
| Effective range | ±0 |
| Long range | −15 |

### Reloading

When a weapon runs out of ammo, it will click and refuse to fire. Press **Shift+R** to open the reload menu. Select your weapon on the left and compatible ammo on the right, then click Reload.

## Melee Combat

Moving into an enemy attacks them. Melee uses HTCS instead of CS. Some melee skills add special attacks bound to specific keys (see [Controls](controls.md)).

Damage from bare hands scales with your **Physique (PH)**. Wielding a melee weapon replaces this with the weapon's own Penetration value.

## Damage and Armor

Every hit lands on a specific body part (head, torso, arm, leg, etc.). Each part has its own HP pool. Armor worn on that part absorbs incoming damage first.

```
Damage dealt = Penetration − Stopping Power
```

- **Penetration** comes from the weapon.
- **Stopping Power (SP)** comes from armor equipped on the hit body part, plus any natural toughness from skills.

If Penetration doesn't exceed Stopping Power, the hit deals no damage.

### Body Part Consequences

| Outcome | Effect |
|---|---|
| Part HP reaches 0 | Part is **broken** — may cause additional penalties |
| Amputated part | Cannot be targeted; removed from hit location pool |
| Vital part broken | May be **lethal** (depends on the entity) |

## Stats That Affect Combat

| Stat | Effect |
|---|---|
| CS | Ranged attack accuracy |
| HTCS | Melee attack accuracy and parry |
| PH | Bare-hands penetration |
| CL (Cool) | Resistance to status effects (poison, slow, etc.) |
