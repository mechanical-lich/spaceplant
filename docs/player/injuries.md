# Injuries & Penalties

Body parts are damaged individually. Taking enough damage to a part has escalating consequences depending on severity.

## Wound Severity

When a body part takes damage, its HP is reduced. As HP drops, the part imposes penalties that apply immediately and persist until you are healed.

### Combat Penalties (Arms, Head, Torso)

Wounds to your arms, head, or torso reduce your attack accuracy. The worst-wounded eligible part determines the penalty — it does not stack across multiple parts.

| Part HP remaining | CS / HTCS Penalty |
|---|---|
| > 75% | None |
| 50 – 75% | −5 |
| 25 – 50% | −15 |
| < 25% | −25 |

The penalty applies to whichever skill governs your attack — **CS** for ranged, **HTCS** for melee. If you are attacking with a wound penalty, the message log will note it: `(wounded: -N)`.

### Movement Penalties (Legs, Feet)

Leg and foot wounds slow your movement by increasing the action point cost to move.

| Condition | Extra cost per affected part |
|---|---|
| Below 50% HP | Movement cost doubled (whole body, not per-part) |
| Broken | +50% of base move cost |
| Amputated | +100% of base move cost |

These stack. A character with one broken leg and one amputated leg pays the doubled cost from the wound, plus +50 for the broken leg, plus +100 for the amputated leg.

## Part States

| State | Meaning |
|---|---|
| **Damaged** | HP reduced but above 0. Wound penalties apply based on severity. |
| **Broken** | HP reached 0. Movement penalties apply for legs. Some vital parts (head, torso) are lethal when broken. |
| **Amputated** | Part destroyed by overkill damage. Removed from the hit location pool — cannot be targeted again. Lethal for some vital parts. |

## Healing

Healing distributes HP evenly across all damaged (non-amputated) parts. Broken parts that recover above 0 HP are no longer broken. Amputated parts cannot be restored.

## Parrying

In melee, you automatically attempt to parry incoming attacks. A successful parry rolls **HTCS + Agility modifier** and halves the damage taken. A wound penalty does not affect your parry roll directly, but a broken arm removes it from the hit location pool, which may affect how future hits land.
