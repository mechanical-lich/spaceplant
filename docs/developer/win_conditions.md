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
| `ArmSelfDestruct` | Player activated the self-destruct console |
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
**Room:** `self_destruct_room` (guaranteed to spawn once on Z=0 via `ThemeEngineering.RequiredRooms`)

**Flow:**
1. Player bumps console → `arm_self_destruct` trigger → `ArmSelfDestructEventData{Turns: 60}`
2. `MainSimState.HandleEvent` sets `sim.SelfDestructTurns` and `sim.selfDestructArmed = true`
3. Each `phaseTurnComplete` decrements `SelfDestructTurns`; at 0, adds `DeadComponent` to player
4. HUD shows countdown in red/amber (`internal/game/mainStateViews.go`)
5. If player reaches a life pod before countdown hits 0 → win via `LifePodEscape` path

**SimWorld fields:**
```go
SelfDestructTurns int  // exported for HUD reads
selfDestructArmed bool // unexported; set by HandleEvent
```

---

## Mother Plant — Two-Phase Entity

### Phase 1: Mobile (`mobile_mother_plant`)

- Spawned at game start (unless Saboteur background) or placed by Saboteur via `PlaceMotherPlantAction`
- Has `MotherPlantSeedComponent{TurnsLeft: 10}`
- Uses `AdvancedAIComponent` with randomness to wander
- After 10 turns, `MotherPlantSeedSystem` kills it and spawns `mother_plant` in its place

### Phase 2: Rooted (`mother_plant`)

- 64×96 sprite, immobile, high HP (200 core + 80 root mass)
- Has `massive_spread_overgrowth` and `immobile` skills
- Death triggers the `extermination` win condition

### `MotherPlantSeedSystem` (`internal/system/MotherPlantSeedSystem.go`)

Implements `ecs.SystemInterface`. Requires `MotherPlantSeedComponent`. `UpdateEntity` fires when the entity has `TurnTaken`, decrements `TurnsLeft`, and transforms at zero:

```go
func (s MotherPlantSeedSystem) Requires() []ecs.ComponentType {
    return []ecs.ComponentType{component.MotherPlantSeed}
}
```

Registered in `SimWorld.NewSimWorld()` before `LightSystem` so it runs before `CleanUpSystem` removes `TurnTaken`.

### Duplicate prevention

`SimWorld.motherPlantExists()` scans all entities for a live `mother_plant` or `mobile_mother_plant` before spawning. `SpawnPlayer` calls this before `spawnMotherPlant(0)` — prevents duplicates when a new player starts on a station where a mother plant is already present.

---

## Life Pod Bay

**Blueprint:** `life_pod_console` in `data/blueprints/environment/environment.json`  
**Room:** `life_pod_bay` (guaranteed to spawn once on Z=0 via `ThemeEngineering.RequiredRooms`)

Interaction triggers:
1. `post_message` — flavor text
2. `life_pod_escape` — handled by `InteractionListener` → emits `LifePodEscapeEventData`

---

## Guaranteed Room Placement

Win-condition rooms must always be present. `FloorTheme.RequiredRooms []string` lists tags that `PlaceRooms` places first (before the weighted random pass) using the shuffled candidate list:

```go
// internal/generation/floor_theme.go
var ThemeEngineering = FloorTheme{
    RequiredRooms: []string{"life_pod_bay", "self_destruct_room"},
    ...
}
```

`PlaceRooms` (`internal/generation/place_rooms.go`) iterates required tags first, consuming candidates until each one fits, then hands remaining candidates to the weighted random loop.

---

## Adding a New Win Condition

1. Emit `eventsystem.GameWonEventData{Outcome: "my_outcome", Message: "..."}` from any system, action, or listener.
2. The win screen picks it up automatically — no other client plumbing needed.
3. If you need a new interaction trigger type, add a case to `listeners/interaction.go`.
4. If you need a guaranteed room, add its tag to `FloorTheme.RequiredRooms` and register a room generator and furniture entry.
5. Add the outcome string to player-facing docs.
