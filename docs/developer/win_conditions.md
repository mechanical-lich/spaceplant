# Win Conditions — Developer Guide

Win conditions are per-player-run events that trigger the win screen, archive the player run, and return to the title. The station always persists after a win.

---

## Architecture

Win conditions emit a single event. Any system, action, listener, or interaction trigger can fire one:

```go
eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{
    Outcome: "my_outcome",
    Message: "Shown on the win screen.",
})
```

### Event types (`internal/eventsystem/eventsystem.go`)

| Constant | Purpose |
|----------|---------|
| `GameWon` | Player won — triggers win screen |
| `LifePodEscape` | Player activated a life pod console |
| `PlaceMotherPlant` | Saboteur placed their cutting |

### Win screen (`internal/game/win_modal.go`)

`WinModal` mirrors `DeathModal`. It registers a `gameWonListener` on `eventsystem.GameWon`. When the event fires, `Show(outcome, message)` is called. The "Return to Title" button calls `OnReturnToTitle`:

```go
s.winModal.OnReturnToTitle = func() {
    SaveStation(s.sim, "saves")
    GraveyardWonPlayerRun("saves", s.sim.PlayerRunID, s.winModal.Outcome)
    s.returnToTitle = true
}
```

Terminal mode (`cmd/terminal/main.go`) uses a package-level `termWinCapture` listener that sets a flag checked in `OnTick`. On win it saves, graveyards, and sends `QuitCommand`.

### Save on win (`internal/game/save.go`)

```go
GraveyardWonPlayerRun(savesDir, playerRunID, outcome string) error
```

Reads the live player save, sets `Won: true`, `Outcome: outcome`, writes to `saves/graveyard/<id>.json`, deletes the live file. The station is saved separately via `SaveStation`.

`PlayerRunMeta` and `PlayerSaveFile` both carry `Won bool` and `Outcome string`.

---

## Implemented Outcomes

### `escape_selfish` — Life Pod Escape

**Trigger:** Player bumps `life_pod_console` → `life_pod_escape` interaction trigger → `LifePodEscapeEventData` → handled in `MainSimState.HandleEvent`.

**Logic (`internal/game/sim_state.go`):**
```go
case eventsystem.LifePodEscape:
    outcome := "escape_selfish"
    msg := "You escape the station alone."
    if skill.HasSkill(player, "saboteur_instinct") && s.sim.MotherPlantPlaced {
        outcome = "saboteur"
        msg = "You've been paid. The station burns behind you."
    }
    eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{...})
```

### `saboteur` — Saboteur Escape

Same trigger as above. If the player has `saboteur_instinct` and `sim.MotherPlantPlaced == true`, the outcome is `"saboteur"` instead of `"escape_selfish"`.

`MotherPlantPlaced` is set to `true` in `MainSimState.HandleEvent` when `PlaceMotherPlantEventData` fires and the `mobile_mother_plant` entity is successfully spawned.

### `extermination` — Kill the Mother Plant

**Trigger:** `CleanUpSystem.Update` detects `entity.Blueprint == "mother_plant"` with `DeadComponent` present AND `SolidComponent` still present (first cleanup frame sentinel).

```go
// internal/system/CleanUpSystem.go
if entity.Blueprint == "mother_plant" && entity.HasComponent(rlcomponents.Solid) {
    eventsystem.EventManager.SendEvent(eventsystem.GameWonEventData{
        Outcome: "extermination",
        Message: "The mother plant is dead. The infestation collapses.",
    })
}
```

This fires regardless of line-of-sight (unlike `DeathEvent`, which is LOS-gated). The `DeathListener` in `listeners/death.go` also has a check, but the `CleanUpSystem` path is the reliable one.

---

## Self-Destruct System

**Blueprint:** `self_destruct_console` in `data/blueprints/environment/environment.json`  
**Room:** `self_destruct_room` (guaranteed to spawn once on Z=5 via `ThemeCommand.RequiredRoomCounts`)  
**Script:** `data/scripts/self_destruct_console.basic`

The console is a live (non-inanimate) entity with an energy budget so it can tick every turn. All arming and countdown logic lives in its script; `SimWorld` and `MainSimState` hold no self-destruct state.

**Flow:**
1. Player interacts with console → `call_script_interact` trigger → `system.CallScriptEvent("on_interact", ...)`.
2. Script's `on_interact`: if not already armed, sets `armed = 1` in `Vars`, sets `Level.Flags["self_destruct_armed"] = 1` and `Level.Flags["self_destruct_turns"] = 60`, posts a warning message.
3. Each turn the console's `on_turn_taken` fires: decrements `turns_left` in `Vars` and syncs `Level.Flags["self_destruct_turns"]`. Posts warnings at ≤10 turns. At zero, calls `kill_player()`.
4. HUD reads `Level.Flags` via `SimWorld.SelfDestructArmed()` and `SimWorld.SelfDestructTurns()` and shows the countdown in amber/red.
5. If the player reaches a life pod before the countdown hits 0 → win via `LifePodEscape` path.

**SimWorld helpers (read-only, no stored state):**
```go
func (sw *SimWorld) SelfDestructArmed() bool { ... } // reads Level.Flags["self_destruct_armed"]
func (sw *SimWorld) SelfDestructTurns() int  { ... } // reads Level.Flags["self_destruct_turns"]
```

**Win condition — heroic death:**

The `plantz` scenario includes a `player_death` rule that fires a win when `self_destruct_armed` is set, letting the player die in the explosion as a valid win state:

```json
{
  "id": "heroic_death",
  "trigger": "player_death",
  "when": [{ "game_flag": "self_destruct_armed" }],
  "result": "win",
  "outcome": "heroic_death"
}
```

See [Scripting](scripting.md) for how the console script works.

---

## Mother Plant — Two-Phase Entity

### Phase 1: Mobile (`mobile_mother_plant`)

- Spawned at game start (unless Saboteur background) or placed by Saboteur via `PlaceMotherPlantAction`
- Has `ScriptComponent` pointing to `data/scripts/mobile_mother_plant.basic` with `turns_left: 10`
- Uses `AdvancedAIComponent` with randomness to wander
- After 10 turns, the script kills it and spawns `mother_plant` in its place

### Phase 2: Rooted (`mother_plant`)

- 64×96 sprite, immobile, high HP (200 core + 80 root mass)
- Has `massive_spread_overgrowth` and `immobile` skills
- Death triggers the `extermination` win condition

### Script (`data/scripts/mobile_mother_plant.basic`)

Handled by the general `ScriptSystem`. Each turn, `on_turn_taken` decrements `turns_left`. At zero it calls `add_dead()` on the cutting and `spawn_entity("mother_plant", x, y, z)` to place the rooted form. See [Scripting](scripting.md).

### Duplicate prevention

`SimWorld.motherPlantExists()` scans all entities for a live `mother_plant` or `mobile_mother_plant` before spawning. `SpawnPlayer` calls this before `spawnMotherPlant(0)` — prevents duplicates when a new player starts on a station where a mother plant is already present.

---

## Life Pod Bay

**Blueprint:** `life_pod_console` in `data/blueprints/environment/environment.json`  
**Room:** `life_pod_bay` (guaranteed to spawn once on Z=2 via `ThemeHabitation.RequiredRoomCounts`)

Interaction triggers:
1. `post_message` — flavor text
2. `life_pod_escape` — handled by `InteractionListener` → emits `LifePodEscapeEventData`

---

## Guaranteed Room Placement

Win-condition rooms must always be present. `FloorTheme.RequiredRoomCounts map[string]int` is stamped onto themes at startup by `applyStationConfig` based on `stationconfig.Config`. The layout of each floor matters: rooms are placed on floors with enough candidate positions.

| Room | Floor | Theme |
|------|-------|-------|
| `life_pod_bay` | Z=2 | Habitation (Grid layout) |
| `self_destruct_room` | Z=5 | Operations & Command (RingSpokes layout) |

`PlaceRooms` (`internal/generation/place_rooms.go`) processes required tags first, consuming candidates until each fits, then runs the weighted random pass for the remainder.

---

## Adding a New Win Condition

1. Emit `eventsystem.GameWonEventData{Outcome: "my_outcome", Message: "..."}` from any system, action, or listener.
2. The win screen picks it up automatically — no other client plumbing needed.
3. If you need a new interaction trigger type, add a case to `listeners/interaction.go`.
4. If you need a guaranteed room, add its tag to `FloorTheme.RequiredRooms` and register a room generator and furniture entry.
5. Add the outcome string to player-facing docs.
