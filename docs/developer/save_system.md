# Save System

Space Plants uses a CDDA-style save structure that separates stations from player runs. This lets stations persist across player deaths and supports multiple concurrent players on the same station.

## Directory Layout

```
saves/
  stations/
    <station-id>/
      station.json    # tiles, non-player entities, floor results, time
      meta.json       # lightweight: id, name, created date
  players/
    <player-run-id>.json   # active (alive) player runs
  graveyard/
    <player-run-id>.json   # dead player runs (moved here on death)
```

IDs are 8-byte random hex strings generated via `crypto/rand`.

## Key Types

```go
// internal/game/save.go

type StationMeta struct {
    StationID string
    Name      string
    Created   time.Time
}

type PlayerRunMeta struct {
    PlayerRunID string
    StationID   string
    Name        string
    ClassName   string
    Dead        bool
    CurrentZ    int
}

type StationSaveFile struct { ... }   // version, tiles, entities, floor results, time
type PlayerSaveFile  struct { ... }   // version, player entity + inventory, seen, counters
```

## SimWorld Identity Fields

`SimWorld` carries three identity strings set at runtime:

| Field | Set by | Meaning |
|-------|--------|---------|
| `StationID` | `RegenerateLevel()` | Unique ID for this station |
| `StationName` | title screen / load | Human-readable station name |
| `PlayerRunID` | `SpawnPlayer()` | Unique ID for this player run |

## Save Functions

| Function | What it saves |
|----------|--------------|
| `SaveStation(sw, dir)` | All entities except the player and player inventory; tiles; floor results |
| `SavePlayerRun(sw, dir)` | Player entity + inventory; seen array; tick/turn counters |
| `SaveAll(sw, dir)` | Calls both of the above |

### Entity Separation

During `SaveStation`, the player entity and anything reachable through the player's `BodyInventory` or `Inventory` components is excluded. During `SavePlayerRun`, only those entities are included.

## Load Functions

| Function | What it loads |
|----------|--------------|
| `LoadStationIntoSimWorld(sw, stationID, dir)` | Tiles + station entities; no player |
| `LoadPlayerRunIntoSimWorld(sw, playerRunID, dir)` | Player entity + inventory; restores seen/counters |
| `LoadFullGame(sw, stationID, playerRunID, dir)` | Calls both of the above |

Deserialization uses a three-pass approach (shared by both load paths):
1. **Stub pass** — create all entities by blueprint ID
2. **Component pass** — restore components from JSON
3. **Inventory pass** — re-link parent→child entity references

## Permadeath Flow

When the player dies:

1. `DeadComponent` is added by the combat system.
2. The client detects it and shows the death modal (graphical) or quits immediately (terminal).
3. `sim.ConvertPlayerToCorpse()` — removes `PlayerComponent` from the entity; sets `sim.Player = nil`. The entity stays in the level as a normal NPC-like entity.
4. `SaveStation(sim, "saves")` — saves the station including the corpse.
5. `GraveyardPlayerRun("saves", playerRunID)` — reads `saves/players/<id>.json`, sets `Dead: true`, writes it to `saves/graveyard/<id>.json`, and deletes the live file.

The player browser only lists `saves/players/`, so the dead run disappears from the UI. The station browser still lists the station; a new player can start there and potentially find the old corpse.

## Auto-Save

Both clients (graphical and terminal) auto-save every 5 player turns using a background goroutine (`go SaveAll(...)`). The pause menu also has an explicit save option.

## Listing Stations and Players

```go
stations, err := ListStations("saves")           // reads saves/stations/*/meta.json
players,  err := ListPlayerRuns("saves", stationID) // reads saves/players/*.json, filters by stationID
```

`ListPlayerRuns` does not read `saves/graveyard/` — dead runs are not presented to the player.

## Adding New Fields to Save Files

1. Add the field to `StationSaveFile` or `PlayerSaveFile` in `save.go`.
2. Populate it in `SaveStation` or `SavePlayerRun`.
3. Read it in `LoadStationIntoSimWorld` or `LoadPlayerRunIntoSimWorld`.
4. Bump `saveVersion` if the change is breaking.

There is no migration path for old saves — this is a pre-release project.
