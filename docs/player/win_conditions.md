# Win Conditions

Space Plants has several ways to win. All of them end your player run — the station persists, but your character's story is over. Won runs are archived in the graveyard alongside dead ones.

---

## Escape via Life Pod

Find the **life pod bay** on the engineering floor (the bottom floor, Z=0). Bump into the **Life Pod Console** to activate it.

> "Emergency escape sequence initiated. Pod away."

You escape the station. The outcome recorded in your graveyard entry depends on what you did before leaving:

| Condition | Outcome |
|-----------|---------|
| Left without doing anything else | `escape_selfish` |
| You are a Saboteur and placed your mother plant cutting | `saboteur` |

---

## Self-Destruct + Escape

The engineering floor also contains a **Self-Destruct Room** with a **Self-Destruct Console**.

1. Bump the console to arm the sequence. A **60-turn countdown** begins. The timer is shown in the HUD.
2. Reach a life pod and escape before the timer hits zero.

If you escape in time, your run is recorded as a win. If the timer reaches zero while you are still on the station, you die.

> The self-destruct and life pod bay may be on opposite ends of the floor — plan your route.

---

## Extermination — Kill the Mother Plant

The source of the infestation is the **Mother Plant**. It starts as a small mobile cutting that wanders the station for about 10 turns before taking root and growing into a massive immobile form.

Kill the large Mother Plant and the infestation collapses. You win immediately.

**Finding it:** The cutting is placed at the start of your run. Use the environment — the infestation tends to spread from where the plant rooted. You do not need line-of-sight for the win to trigger.

**Killing it early:** If you find and kill the cutting before it roots, it just dies — no win. The large rooted form is the win condition.

---

## Saboteur Path

The **Saboteur** background gives you the `Saboteur Instinct` skill, which binds the **Place Mother Plant** action to `X`.

Your goal: introduce the infestation yourself, then escape.

1. Use `X` to place your mother plant cutting anywhere on the station.
2. The cutting will wander and root within 10 turns.
3. Reach the life pod bay and escape.

If you escape after placing the cutting, your outcome is recorded as `saboteur` — a clean contract.

> If any crew has line-of-sight when you place the cutting, they turn hostile. Place it somewhere private.

---

## After Winning

- Your player run is archived in the graveyard with `Won: true`.
- The station is saved intact — other players can start runs there and find the consequences of your actions (a rooted mother plant, an empty life pod bay, etc.).
- Your character no longer appears in the player list for that station.
