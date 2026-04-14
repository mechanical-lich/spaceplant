# Floor Generators

Floor generation is orchestrated by `generateFloor` in [floor_generators.go](../../internal/generation/floor_generators.go), which dispatches to a layout-specific generator based on the floor's `FloorTheme.Layout` field. Each generator carves the level geometry, buds smaller rooms off open walls, places maintenance tunnels, and returns a tagged `[]Room` slice.

## Core Concepts

**Carving** — All generators write tiles directly into the `world.Level` grid using three primitives from [shapes.go](../../internal/generation/shapes.go):

| Function | What it draws |
|---|---|
| `CarveRoom(x, y, w, h, ...)` | Filled rectangle: wall border, floor interior |
| `CarveRect(x1, y1, x2, y2, ...)` | Same as CarveRoom but specified by two corners |
| `CarveCircle(cx, cy, r, ...)` | Filled circle: wall ring, floor interior |

Both `CarveRoom` and `CarveCircle` accept two flags:
- `noOverwrite` — skip tiles that are already floor or wall (won't clobber existing rooms).
- `noBudding` — mark tiles as ineligible for `BudRooms` to grow a side room off of.

**BudRooms** — After the structural skeleton is carved, `BudRooms` walks all open wall tiles and randomly grows small rectangular rooms off them up to `theme.BudCount` attempts. These are the main source of tagged `Room` values returned to the caller.

**Maintenance tunnels** — `CarveMaintenanceTunnels` scatters narrow 1-tile corridors across the floor to ensure connectivity. The density argument controls how many are attempted.

**Doors** — `spawnDoor` converts a specific wall tile into a passable door. `flushDoors` converts any wall tile that borders two floor regions into a door, closing off structural seams.

**Polish** — `l.Polish(z)` smooths rough tile edges after carving.

**Room tagging** — `tagRooms` assigns a theme-weighted `Tag` string (e.g. `"crew_quarters"`) to each returned `Room` via `FloorTheme.pickRoomTag()`.

---

## Layouts

### `ring_spokes` — Ring Spokes

**Used by:** Science & Research, Operations & Command

**Structure:**
1. A circular hub (`CarveCircle`) centered on the floor, radius ≈ `min(W,H)/8`. Marked `noBudding` so rooms don't grow directly out of the hub.
2. Four rectangular hallway arms extending from the hub to the floor edges — one per cardinal direction. Width of each arm is randomized `[5,10]`. Arms punch through the hub wall (`noOverwrite=false`) so the boundary becomes passable.
3. Doors placed at the four hub/arm junctions.
4. `BudRooms` grows side rooms off the arms.
5. Maintenance tunnels at density 15.

**Expected output:** A recognizable plus/asterisk shape. The hub is a large open circle; four corridors radiate outward. Side rooms cluster along the corridors. Feels deliberate and organized — good for command and science floors where compartmentalization matters.

---

### `grid` — Grid

**Used by:** Habitation, Commerce & Social

**Structure:**
1. One large central hallway, randomly oriented wide or tall (`[W/2, W*3/4]` long, `[5,10]` in the short axis).
2. A second hallway perpendicular to the first, same size range. The two halls form a `+` shape. The second is carved with `noOverwrite=true` so it doesn't erase the first's walls where they cross.
3. Both hall interiors are re-carved as pure floor to fix any wall tiles left at the intersection.
4. `BudRooms` grows many small rooms off both halls (high `BudCount`: 80–100).
5. Polish → maintenance tunnels (density 10) → polish.

**Expected output:** A cross of two wide corridors with many small rooms budded off them. High room density makes it feel dense and lived-in — ideal for habitation and commercial spaces.

---

### `industrial_ring` — Industrial Ring

**Used by:** Engineering & Systems

**Structure:**
1. Outer circle, radius ≈ `min(W,H)/4`.
2. Inner circle, radius ≈ outer/2, carved with `noBudding=true` — this hollows out the center leaving a ring corridor between the two circles.
3. Four spoke corridors (width `[3,5]`) connecting the inner ring wall to the outer ring wall at N/S/E/W. Spokes are positioned to overlap the inner ring wall tile so the junction is open.
4. Doors placed on the inner ring wall at each spoke entry.
5. `BudRooms` grows rooms off the ring and spokes.
6. Polish → maintenance tunnels (density 30, highest of all layouts) → polish.

**Expected output:** Two concentric rings connected by four spokes. The inner ring is a navigable loop; the outer ring is the main travel corridor. Heavy maintenance tunnel density mirrors an engineering floor's need for service access everywhere.

---

### `open_bays` — Open Bays

**Used by:** Logistics & Industry

**Structure:**
1. 3–5 large rectangular bays (`W/4 × H/4` ± 4 tiles), placed at five fixed anchor positions: four near-corners and one center. The number of bays is randomized `[3,5]`.
2. Bays are connected in sequence by L-shaped wide corridors (5 tiles wide): a horizontal leg between bay centers then a vertical leg.
3. Explicit bays are immediately tagged with `theme.pickRoomTag()`.
4. A smaller set of `BudRooms` (low `BudCount`: 40) adds utility side rooms off the corridors.
5. Polish → maintenance tunnels (density 20) → polish.

**Expected output:** A handful of large open rectangular rooms connected by wide corridors — appropriate for cargo and freight operations. Much less room density than Grid; bays dominate the floor.

---

### `rectangle` — Rectangle

**Used by:** Mixed-use floors

**Structure:**
1. Four corner anchor rooms (`W/6 × H/6`), one at each corner. Marked `noBudding=true`.
2. Four connecting hallways (5 tiles wide) along the edges, punching through corner room walls so corridors and rooms share a passable tile.
3. Eight doors placed at every corridor/corner-room junction (two per corner room).
4. A circular room (`CarveCircle`) centered on the floor, radius `W/8`, as a central hub. Marked `noBudding=true`.
5. Corner rooms are tagged explicitly; `BudRooms` adds side rooms off the hallways.
6. Polish → maintenance tunnels (density 20) → polish.

**Expected output:** Four rooms anchoring the corners, connected by edge-hugging corridors, with a circular room in the center. The combination of dedicated corners, halls, and a hub gives this layout a mixed-use feel.

---

## Return Value

All generators return `[]Room`. Each `Room` is:

```go
type Room struct {
    X, Y, Width, Height int
    Tag                 string // e.g. "crew_quarters", "reactor_core"
}
```

Explicitly carved rooms (bays, corner rooms) are included directly. `BudRooms`-generated rooms are appended after. The full slice is used downstream by `populate.go` to spawn furniture, NPCs, and items appropriate to each room's tag.

---

## Adding a New Layout

1. Add a `Layout*` constant to [floor_theme.go](../../internal/generation/floor_theme.go).
2. Write a `generateMyLayout(l, z, theme) []Room` function in [floor_generators.go](../../internal/generation/floor_generators.go) following the pattern above: carve structure → `BudRooms` → optional polish → `CarveMaintenanceTunnels` → `flushDoors` → `tagRooms` → return.
3. Add the new case to the `switch` in `generateFloor`.
4. Create a `FloorTheme` var that references the new layout constant and add it to `FloorStack` if it should be part of the default station.
