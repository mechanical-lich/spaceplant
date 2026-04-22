# Developer Console — Guide

The developer console is a cheat/debug modal available in-game. Open it with **Shift+ESC**. Type a command and press OK (or Enter).

---

## Commands

### `teleport <x> <y> <z>` / `tp <x> <y> <z>`

Moves the player to tile coordinates `(x, y, z)` instantly. The camera follows. The spatial index is updated correctly so the player sprite renders at the new position.

```
tp 50 30 5
teleport 100 100 0
```

---

### `heal [amount]`

Restores HP to the player. Defaults to 50 if no amount is given.

```
heal
heal 200
```

---

### `find <blueprint>`

Searches all live and static entities on every floor for the nearest instance of the given blueprint (case-insensitive). Reports its coordinates to the message log. Useful for locating win-condition entities like `self_destruct_console` or `mother_plant` without exploring.

Distance is measured using a squared offset metric (`dx*dx + dy*dy`) with an additional 50× penalty per floor difference, so entities on the current floor are strongly preferred.

```
find self_destruct_console
find mother_plant
find life_pod_console
```

If no match is found, the message log reports `[cheat] no "<blueprint>" found`.

---

### `find_room <tag>` / `fr <tag>`

Searches all floor results for the nearest room whose tag matches (case-insensitive). Reports the room's origin coordinates and floor to the message log.

Distance uses the same formula as `find` — a squared XY offset metric (`dx*dx + dy*dy`) plus a 50× floor-difference penalty.

```
find_room self_destruct_room
fr life_pod_bay
fr crew_quarters
```

---

## Implementation

`internal/game/cheat_modal.go` — `CheatModal.runCommand`. Adding a new command is a `case` in the `switch` on `parts[0]`.
