package game

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlasciiclient"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// asciiMap maps entity blueprint names to ASCII display characters.
var asciiMap = map[string]asciiGlyph{
	"player":  {"@", 0, 255, 0},
	"enemy":   {"E", 255, 0, 0},
	"rat":     {"r", 200, 150, 100},
	"health":  {"+", 255, 50, 50},
	"sword":   {"/", 200, 200, 200},
	"default": {"?", 150, 150, 150},
}

// tileAsciiMap maps tile type indices to ASCII glyphs (at full visibility).
var tileAsciiMap map[int]asciiGlyph

func buildTileAsciiMap() map[int]asciiGlyph {
	return map[int]asciiGlyph{
		world.TypeWall:                   {"#", 180, 180, 180},
		world.TypeFloor:                  {".", 80, 80, 80},
		world.TypeDoor:                   {"+", 200, 180, 50},
		world.TypeStairsUp:               {"<", 100, 200, 200},
		world.TypeStairsDown:             {">", 100, 200, 200},
		world.TypeMaintenanceTunnelWall:  {"#", 120, 100, 80},
		world.TypeMaintenanceTunnelFloor: {".", 60, 50, 40},
		world.TypeMaintenanceTunnelDoor:  {"+", 160, 130, 40},
	}
}

type asciiGlyph struct {
	char    string
	r, g, b uint8
}

// dim returns a darkened version of the glyph for previously-seen tiles.
func (g asciiGlyph) dim() asciiGlyph {
	return asciiGlyph{g.char, g.r / 4, g.g / 4, g.b / 4}
}

// SPCodec encodes tiles and entities as ASCII snapshots for the terminal client,
// applying fog of war: visible tiles are bright, previously-seen tiles are dimmed,
// and never-seen tiles are omitted. Entities are only shown when visible.
type SPCodec struct {
	sim *SimWorld
}

// NewSPCodec creates a codec that encodes tiles from sim's active Z layer.
func NewSPCodec(sim *SimWorld) *SPCodec {
	return &SPCodec{sim: sim}
}

// compile-time assertion
var _ transport.SnapshotCodec = (*SPCodec)(nil)

// Encode builds a snapshot applying fog of war using LOS + seen state.
func (c *SPCodec) Encode(tick uint64, entities []*ecs.Entity) *transport.Snapshot {
	if tileAsciiMap == nil {
		tileAsciiMap = buildTileAsciiMap()
	}

	z := c.sim.CurrentZ
	level := c.sim.Level.Level
	lw, lh := level.GetWidth(), level.GetHeight()

	// Player position for LOS checks. If no player, encode nothing.
	if c.sim.Player == nil || !c.sim.Player.HasComponent("Position") {
		return transport.NewSnapshot(tick, nil)
	}
	pc := c.sim.Player.GetComponent("Position").(*component.PositionComponent)
	px, py := pc.GetX(), pc.GetY()

	snaps := make([]*transport.EntitySnapshot, 0, lw*lh/4+len(entities))

	// Encode tiles with fog of war.
	for x := 0; x < lw; x++ {
		for y := 0; y < lh; y++ {
			tile := level.GetTilePtr(x, y, z)
			if tile == nil {
				continue
			}
			def := rlworld.TileDefinitions[tile.Type]
			if def.Air {
				continue
			}
			glyph, ok := tileAsciiMap[tile.Type]
			if !ok {
				continue
			}

			visible := rlfov.Los(level, px, py, x, y, z)
			if visible {
				c.sim.Level.SetSeen(x, y, z, true)
			} else if !c.sim.Level.GetSeen(x, y, z) {
				continue // never seen — omit
			} else {
				glyph = glyph.dim() // seen before but not currently visible
			}

			snaps = append(snaps, &transport.EntitySnapshot{
				ID:        fmt.Sprintf("t:%d,%d,%d", x, y, z),
				Blueprint: "tile",
				Components: map[ecs.ComponentType]transport.ComponentData{
					rlcomponents.AsciiAppearance: &rlcomponents.AsciiAppearanceComponent{
						Character: glyph.char,
						R:         glyph.r,
						G:         glyph.g,
						B:         glyph.b,
					},
					rlcomponents.Position: &rlcomponents.PositionComponent{
						X: x, Y: y, Z: z,
					},
				},
			})
		}
	}

	// Encode entities — only those currently visible to the player.
	for _, e := range entities {
		if e == nil {
			continue
		}
		if !e.HasComponent("Position") {
			continue
		}
		epc := e.GetComponent("Position").(*component.PositionComponent)
		if epc.GetZ() != z {
			continue
		}
		if !rlfov.Los(level, px, py, epc.GetX(), epc.GetY(), z) {
			continue
		}
		var glyph asciiGlyph
		if e.HasComponent(rlcomponents.AsciiAppearance) {
			ac := e.GetComponent(rlcomponents.AsciiAppearance).(*rlcomponents.AsciiAppearanceComponent)
			glyph = asciiGlyph{string(ac.Character), ac.R, ac.G, ac.B}
		} else {
			var ok bool
			glyph, ok = asciiMap[e.Blueprint]
			if !ok {
				glyph = asciiMap["default"]
			}
		}
		snaps = append(snaps, &transport.EntitySnapshot{
			ID:        fmt.Sprintf("%p", e),
			Blueprint: e.Blueprint,
			Components: map[ecs.ComponentType]transport.ComponentData{
				rlcomponents.AsciiAppearance: &rlcomponents.AsciiAppearanceComponent{
					Character: glyph.char,
					R:         glyph.r,
					G:         glyph.g,
					B:         glyph.b,
				},
				rlcomponents.Position: &rlcomponents.PositionComponent{
					X: epc.GetX(),
					Y: epc.GetY(),
					Z: epc.GetZ(),
				},
			},
		})
	}

	return transport.NewSnapshot(tick, snaps)
}

// Decode applies the snapshot to an *rlasciiclient.AsciiWorld.
func (c *SPCodec) Decode(snap *transport.Snapshot, world any) {
	w := world.(*rlasciiclient.AsciiWorld)
	alive := make(map[string]bool, len(snap.Entities))
	for _, es := range snap.Entities {
		alive[es.ID] = true
		e := w.FindOrCreate(es.ID, es.Blueprint)
		if raw, ok := es.Components[rlcomponents.AsciiAppearance]; ok {
			e.AddComponent(raw.(*rlcomponents.AsciiAppearanceComponent))
		}
		if raw, ok := es.Components[rlcomponents.Position]; ok {
			rpc := raw.(*rlcomponents.PositionComponent)
			setAsciiPos(e, rpc.X, rpc.Y, rpc.Z)
		}
	}
	w.RemoveNotIn(alive)
}

func setAsciiPos(e *ecs.Entity, x, y, z int) {
	if !e.HasComponent(rlcomponents.Position) {
		e.AddComponent(&rlcomponents.PositionComponent{X: x, Y: y, Z: z})
		return
	}
	ep := e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	ep.X, ep.Y, ep.Z = x, y, z
}
