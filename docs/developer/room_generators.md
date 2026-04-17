# Room Generators

Room generators run after floor layout carving and before population. They sculpt the interior of a tagged room (optional sub-walls, platform rings, etc.) and return **placement hints** that tell the populate pass where to put furniture.

Implemented in [room_generators.go](../../internal/generation/room_generators.go).

---

## Concepts

### PlacementRegion

A `PlacementRegion` constrains where furniture lands inside a room:

| Constant | Meaning |
|---|---|
| `RegionAnywhere` | No constraint — pick any open floor tile |
| `RegionNorthWall` | Row of tiles adjacent to the north interior wall |
| `RegionSouthWall` | Row adjacent to south interior wall |
| `RegionEastWall` | Column adjacent to east interior wall |
| `RegionWestWall` | Column adjacent to west interior wall |
| `RegionCenter` | Middle third of both axes |

### PlacementHint

```go
type PlacementHint struct {
    Blueprint string
    Region    PlacementRegion
}
```

A hint says "place this blueprint in this region." The populate pass respects hints and also skips any tile adjacent to a door.

### DoorDir

Every budded room records `DoorDir [2]int` — the cardinal direction from the hallway *into* the room. Room generators use this to orient furniture:

- `dirToFarWallRegion(doorDir)` → wall opposite the door (good for focal equipment)
- `dirToNearWallRegion(doorDir)` → wall with the door (good for entrance items)
- `doorDirPerpendicularWalls(doorDir)` → the two side walls (good for bench rows)

---

## Room Size Ranges

Each tag has a preferred `RoomSizeRange{MinW, MaxW, MinH, MaxH}` defined in `roomSizes`. `PlaceRooms` calls `RoomSizeFor(tag)` before carving so each room is appropriately sized for its generator. Tags with no entry use `defaultRoomSize = {7,11, 6,9}`.

Size philosophy:
- **Sleeping rooms** (crew_quarters, officers_suite) — wider to fit a partition wall + two zones
- **Hangar/cargo** — largest sizes, open-bay feel
- **Utility/shop/interrogation** — smallest, intentionally cramped
- **Labs** — medium with near-square proportions for bench rows
- **Bridge** — large enough for the inner platform ring

---

## Generator Types

All generators implement `RoomGenerator`:

```go
type RoomGenerator interface {
    Generate(l *world.Level, room Room, z int, doorDir [2]int) []PlacementHint
}
```

### `bridgeGenerator`

Carves an inner rectangular ring (raised platform) with four gap openings at the cardinal midpoints. Places `captains_chair` / `map_table` in the center and consoles on the far wall.

### `sleepingRoomGenerator`

Optionally carves a partition wall 2/3 of the way from the door to the far wall, dividing the room into a sleeping alcove and a living area. Beds go in the far zone; storage and desks near the entrance.

Fields: `hasPartition bool`, `bedBlueprint string`, `farItems []string`, `nearItems []string`.

### `labStyleGenerator`

No geometry carving. Distributes bench/instrument items alternating across the two side walls, leaving a central aisle. Specialty equipment on the far wall.

Fields: `sideItems []string`, `farItems []string`.

### `controlRoomGenerator`

No geometry carving. General-purpose workstation layout: consoles on the far wall, optional center display, optional side panels, optional entrance items.

Fields: `farItems`, `centerItems`, `sideItems`, `nearItems []string`.

### `storageRoomGenerator`

No geometry carving. Distributes wall items round-robin across all four walls; center items go in `RegionCenter`.

Fields: `wallItems []string`, `centerItems []string`.

### `rowBedsGenerator`

No geometry carving. Places pairs of beds alternating on the two side walls (parallel rows). Equipment on the far wall, utility items near the entrance.

Fields: `bedBlueprint string`, `farItems []string`, `nearItems []string`.

### `workshopGenerator`

No geometry carving. Workbenches/racks alternate across side walls; specialty equipment on the far wall; utility items near the entrance.

Fields: `sideItems []string`, `farItems []string`, `nearItems []string`.

### `socialRoomGenerator`

No geometry carving. Tables/seating in the center, service equipment on the far wall, decorative items on sides, staff items near entrance.

Fields: `centerItems`, `farItems`, `sideItems`, `nearItems []string`.

---

## Adding a New Generator

1. Implement the `RoomGenerator` interface.
2. Register it in `roomGenerators` keyed by tag name.
3. Add a `RoomSizeRange` entry in `roomSizes` if the default doesn't fit.

To add a new room tag end-to-end:
1. Add it to the appropriate `FloorTheme.RoomWeights` in [floor_theme.go](../../internal/generation/floor_theme.go).
2. Add its size range to `roomSizes`.
3. Register a generator in `roomGenerators`.

---

## Door Safety

`populateRoom` in [populate.go](../../internal/generation/populate.go) calls `adjacentToDoor` before placing any piece of furniture. Tiles directly adjacent (cardinal) to a door entity are skipped regardless of what the hint requests, ensuring doors are never blocked.
