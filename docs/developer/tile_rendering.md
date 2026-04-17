# Tile Rendering

Tiles are drawn in [drawing.go](../../internal/world/drawing.go) via `DrawTile`. The logical grid is **32×32** pixels per tile, but tile and entity sprites can be larger and offset to overdraw into adjacent rows — useful for walls with visual height, tall characters, etc.

---

## Tile Sprites

`DrawTile` resolves the sprite from the tile's `TileDefinition` (from ml-rogue-lib `rlworld.TileDefinitions`).

### Default behavior

The source rectangle is `spriteX * tileW` wide and `tileH` tall, drawn at the tile's screen position with no offset.

### Per-tile size override

`TileDefinition` supports optional sprite dimensions and offset:

| JSON field | Go field | Default | Effect |
|---|---|---|---|
| `"spriteWidth"` | `SpriteWidth int` | 0 (= tile width) | Width of the source rect in the spritesheet |
| `"spriteHeight"` | `SpriteHeight int` | 0 (= tile height) | Height of the source rect |
| `"spriteOffsetX"` | `SpriteOffsetX int` | 0 | Screen X offset when drawing |
| `"spriteOffsetY"` | `SpriteOffsetY int` | 0 | Screen Y offset when drawing |

Zero values mean "use tile dimensions" — existing tiles with no overrides are unaffected.

**Example** — a wall tile whose sprite is 32×40 (8px taller than the tile grid, bleeding upward):

```json
{
  "name": "stone_wall",
  "solid": true,
  "autoTile": 1,
  "spriteHeight": 40,
  "spriteOffsetY": -8,
  "variants": [...]
}
```

This draws 40px tall from the spritesheet but positions it 8px higher on screen, so the extra 8px bleeds visually into the tile row above — the same technique used by Cataclysm: DDA tilesets for wall depth.

---

## Entity Sprites

Entities use `AppearanceComponent` (or `LayeredAppearanceComponent` for crewed characters):

| Field | Effect |
|---|---|
| `SpriteWidth` / `SpriteHeight` | Total sprite dimensions in the sheet (e.g. 32×48) |
| `SpriteOffsetX` / `SpriteOffsetY` | Screen offset when drawing (e.g. `SpriteOffsetY: -16` for 32×48 on 32×32 tiles) |

Multi-tile entities (`SizeComponent` width/height > 1) subdivide the sprite sheet region: each sub-tile samples `SpriteWidth/entityW × SpriteHeight/entityH` from the appropriate column/row of the sprite.

The standard offset for 32×48 entity sprites on 32×32 tiles is `SpriteOffsetY: -16`. Large entities (64×96, e.g. abomination) use the same -16 offset.

Layered human sprites (`drawLayered`) apply a hardcoded -16 offset regardless of blueprint values.
