package game

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/resource"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/generation"
)

const (
	mapTileSize = 5  // pixels per tile on the station map
	mapPad      = 16 // padding inside the overlay
	mapHeaderH  = 36 // height of title + floor label bar
	mapFooterH  = 40 // height of button row at the bottom
)

// Waypoint is a player-set map marker at a specific world tile.
type Waypoint struct {
	X, Y, Z int
	Active  bool
}

// StationMapModal is a full-screen overlay showing the entire station layout.
// The player can browse floors and click a tile to place a waypoint.
type StationMapModal struct {
	Visible  bool
	sim      *SimWorld
	viewedZ  int
	waypoint *Waypoint // pointer shared with SPClientState

	// Derived each frame
	mapOriginX int // screen X where the map image starts
	mapOriginY int // screen Y where the map image starts
	mapW       int // pixel width of the map render
	mapH       int // pixel height of the map render
}

func newStationMapModal(sim *SimWorld, wp *Waypoint) *StationMapModal {
	return &StationMapModal{sim: sim, waypoint: wp}
}

func (m *StationMapModal) Open() {
	if m.sim.Player != nil {
		pc := m.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
		m.viewedZ = pc.GetZ()
	}
	m.Visible = true
}

func (m *StationMapModal) Update() {
	if !m.Visible {
		return
	}

	cfg := config.Global()
	sw := cfg.ScreenWidth
	sh := cfg.ScreenHeight

	// Left/right arrow or A/D to navigate floors.
	results := m.sim.FloorResults
	numFloors := len(results)
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if m.viewedZ > 0 {
			m.viewedZ--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if m.viewedZ < numFloors-1 {
			m.viewedZ++
		}
	}

	// Compute map area geometry (same calc as Draw uses).
	overlayX := mapPad
	overlayY := mapPad
	overlayW := sw - mapPad*2
	overlayH := sh - mapPad*2
	contentY := overlayY + mapHeaderH
	contentH := overlayH - mapHeaderH - mapFooterH
	contentW := overlayW

	lw := m.sim.Level.Width
	lh := m.sim.Level.Height
	mw := lw * mapTileSize
	mh := lh * mapTileSize

	// Centre the map within the content area.
	mox := overlayX + (contentW-mw)/2
	moy := contentY + (contentH-mh)/2

	m.mapOriginX = mox
	m.mapOriginY = moy
	m.mapW = mw
	m.mapH = mh

	// Left-click inside the map sets a waypoint.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		if cx >= mox && cx < mox+mw && cy >= moy && cy < moy+mh {
			tx := (cx - mox) / mapTileSize
			ty := (cy - moy) / mapTileSize
			// Only allow waypoint on a seen tile.
			if m.sim.Level.GetSeen(tx, ty, m.viewedZ) {
				m.waypoint.X = tx
				m.waypoint.Y = ty
				m.waypoint.Z = m.viewedZ
				m.waypoint.Active = true
			}
		} else {
			// Click outside map — close.
			m.Visible = false
		}
	}

	// Right-click clears waypoint.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		m.waypoint.Active = false
	}
}

func (m *StationMapModal) Draw(screen *ebiten.Image) {
	if !m.Visible {
		return
	}

	cfg := config.Global()
	sw := float64(cfg.ScreenWidth)
	sh := float64(cfg.ScreenHeight)

	// Semi-transparent full-screen backdrop.
	ebitenutil.DrawRect(screen, 0, 0, sw, sh, color.RGBA{0, 0, 0, 200})

	// Overlay panel.
	overlayX := float64(mapPad)
	overlayY := float64(mapPad)
	overlayW := sw - float64(mapPad)*2
	overlayH := sh - float64(mapPad)*2
	ebitenutil.DrawRect(screen, overlayX, overlayY, overlayW, overlayH, color.RGBA{20, 22, 28, 255})
	ebitenutil.DrawRect(screen, overlayX, overlayY, overlayW, 1, color.RGBA{80, 80, 100, 255})
	ebitenutil.DrawRect(screen, overlayX, overlayY+overlayH-1, overlayW, 1, color.RGBA{80, 80, 100, 255})
	ebitenutil.DrawRect(screen, overlayX, overlayY, 1, overlayH, color.RGBA{80, 80, 100, 255})
	ebitenutil.DrawRect(screen, overlayX+overlayW-1, overlayY, 1, overlayH, color.RGBA{80, 80, 100, 255})

	// Header: title and floor name.
	results := m.sim.FloorResults
	floorName := "Unknown"
	if m.viewedZ >= 0 && m.viewedZ < len(results) && results[m.viewedZ].Theme != nil {
		floorName = results[m.viewedZ].Theme.Name
	}
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("STATION MAP  Left and Right Arrows to browse floors   right-click to clear waypoint   click map to place waypoint   ESC to close"),
		int(overlayX)+8, int(overlayY)+6)
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Floor %d/%d — %s", m.viewedZ+1, len(results), floorName),
		int(overlayX)+8, int(overlayY)+18)

	// Map image.
	mapImg := m.renderFloor(m.viewedZ)
	if mapImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(m.mapOriginX), float64(m.mapOriginY))
		screen.DrawImage(mapImg, op)
	}

	// Room number labels drawn directly on screen at room centroids.
	if m.viewedZ >= 0 && m.viewedZ < len(results) {
		for _, room := range results[m.viewedZ].Rooms {
			cx := room.CenterX()
			cy := room.CenterY()
			if !m.sim.Level.GetSeen(cx, cy, m.viewedZ) {
				continue
			}
			sx := m.mapOriginX + cx*mapTileSize - 4
			sy := m.mapOriginY + cy*mapTileSize - 3
			label := fmt.Sprintf("%d", room.Number)
			ebitenutil.DebugPrintAt(screen, label, sx, sy)
		}
	}

	// Waypoint marker on map.
	if m.waypoint.Active && m.waypoint.Z == m.viewedZ {
		wx := float64(m.mapOriginX + m.waypoint.X*mapTileSize)
		wy := float64(m.mapOriginY + m.waypoint.Y*mapTileSize)
		ebitenutil.DrawRect(screen, wx-1, wy-1, float64(mapTileSize)+2, float64(mapTileSize)+2, color.RGBA{255, 220, 0, 255})
		ebitenutil.DrawRect(screen, wx, wy, float64(mapTileSize), float64(mapTileSize), color.RGBA{255, 165, 0, 200})
	}

	// Player position — bright blue dot.
	if m.sim.Player != nil {
		pc := m.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
		if pc.GetZ() == m.viewedZ {
			dotSize := float64(mapTileSize + 4)
			px := float64(m.mapOriginX+pc.GetX()*mapTileSize) - 2
			py := float64(m.mapOriginY+pc.GetY()*mapTileSize) - 2
			ebitenutil.DrawRect(screen, px-1, py-1, dotSize+2, dotSize+2, color.RGBA{255, 255, 255, 255})
			ebitenutil.DrawRect(screen, px, py, dotSize, dotSize, color.RGBA{60, 120, 255, 255})
		}
	}

	// Footer: floor nav hint.
	footerY := int(overlayY) + int(overlayH) - mapFooterH + 12
	ebitenutil.DebugPrintAt(screen, m.floorNavLine(results), int(overlayX)+8, footerY)
}

// renderFloor builds a tile-colored image for floor z. Unseen tiles are dark.
func (m *StationMapModal) renderFloor(z int) *ebiten.Image {
	lw := m.sim.Level.Width
	lh := m.sim.Level.Height
	img := ebiten.NewImage(lw*mapTileSize, lh*mapTileSize)

	cfg := config.Global()
	sw, sh := cfg.TileSizeW, cfg.TileSizeH
	mapTex := resource.Textures["map"]

	for ty := 0; ty < lh; ty++ {
		for tx := 0; tx < lw; tx++ {
			seen := m.sim.Level.GetSeen(tx, ty, z)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(mapTileSize)/float64(sw), float64(mapTileSize)/float64(sh))
			op.GeoM.Translate(float64(tx*mapTileSize), float64(ty*mapTileSize))

			tile := m.sim.Level.Level.GetTilePtr(tx, ty, z)
			if tile == nil || !seen {
				ebitenutil.DrawRect(img, float64(tx*mapTileSize), float64(ty*mapTileSize),
					float64(mapTileSize), float64(mapTileSize), color.RGBA{10, 10, 14, 255})
				if tile != nil {
					// Tile exists but unseen — dim ghost tint.
					variant := m.sim.Level.Level.ResolveVariant(tile)
					spriteX := variant.SpriteX * sw
					sub := mapTex.SubImage(image.Rect(spriteX, variant.SpriteY, spriteX+sw, variant.SpriteY+sh)).(*ebiten.Image)
					op2 := &ebiten.DrawImageOptions{}
					op2.ColorScale.ScaleAlpha(0.15)
					op2.GeoM.Scale(float64(mapTileSize)/float64(sw), float64(mapTileSize)/float64(sh))
					op2.GeoM.Translate(float64(tx*mapTileSize), float64(ty*mapTileSize))
					img.DrawImage(sub, op2)
				}
				continue
			}

			variant := m.sim.Level.Level.ResolveVariant(tile)
			spriteX := variant.SpriteX * sw
			sub := mapTex.SubImage(image.Rect(spriteX, variant.SpriteY, spriteX+sw, variant.SpriteY+sh)).(*ebiten.Image)
			img.DrawImage(sub, op)
		}
	}

	return img
}

// floorNavLine builds the ← Floor 1 [2] Floor 3 → navigation text.
func (m *StationMapModal) floorNavLine(results []generation.FloorResult) string {
	line := ""
	for i, fr := range results {
		if i > 0 {
			line += "  "
		}
		name := ""
		if fr.Theme != nil {
			name = fr.Theme.Name
		}
		if i == m.viewedZ {
			line += fmt.Sprintf("[F%d %s]", i+1, name)
		} else {
			line += fmt.Sprintf("F%d %s", i+1, name)
		}
	}
	return line
}
