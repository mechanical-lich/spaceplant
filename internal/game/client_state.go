package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/client"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/ui"
)

// compile-time assertion
var _ client.ClientState = (*SPClientState)(nil)

// SPClientState is the graphical (Ebiten) client state.
// It polls keyboard input and forwards action commands to the server,
// then renders the level and HUD each frame.
type SPClientState struct {
	sim           *SimWorld
	transport     transport.ClientTransport
	gui           *ui.GUI
	inventoryView *InventoryView
	CameraX       int
	CameraY       int
	pressDelay    int
}

// NewSPClientState creates a ready-to-use graphical client state.
func NewSPClientState(sim *SimWorld, t transport.ClientTransport) *SPClientState {
	cs := &SPClientState{
		sim:           sim,
		transport:     t,
		gui:           ui.NewGUI(&GUIViewMain{}),
		inventoryView: NewInventoryView(sim.Player),
	}
	if sim.Player != nil {
		pc := sim.Player.GetComponent("Position").(*component.PositionComponent)
		cs.CameraX = pc.GetX()
		cs.CameraY = pc.GetY()
	}
	return cs
}

func (s *SPClientState) Done() bool { return false }

// Update is called every Ebiten frame. It handles input, sends commands to the
// server, animates sprites, and updates the HUD.
func (s *SPClientState) Update(_ *transport.Snapshot) client.ClientState {
	mlgeevent.GetQueuedInstance().HandleQueue()

	fps := ebiten.ActualFPS()
	tps := ebiten.ActualTPS()
	s.sim.Mu.RLock()
	turnCount := s.sim.TurnCount
	s.sim.Mu.RUnlock()
	title := fmt.Sprintf("%s - Turn: %d FPS: %.1f TPS: %.1f", "Space Plants!", turnCount, fps, tps)
	ebiten.SetWindowTitle(title)

	// Close inventory on Escape.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && s.inventoryView.Visible {
		s.inventoryView.Visible = false
		return nil
	}

	// Open inventory on I.
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		s.inventoryView.Visible = true
		return nil
	}

	// GUI, inventory, and turn check are read-only — RLock is sufficient.
	s.sim.Mu.RLock()
	s.gui.Update(s)
	s.inventoryView.Update()
	hasTurn := s.sim.Player != nil && s.sim.Player.HasComponent("MyTurn")
	s.sim.Mu.RUnlock()

	// Send movement/action commands only when it is the player's turn.
	if hasTurn {
		// Toggle commands — fire once on key-down, ignore held repeats.
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			s.transport.SendCommand(&transport.Command{
				Type:    CmdAction,
				Payload: ActionPayload{Key: "R"},
			})
		}

		if s.pressDelay > 0 {
			s.pressDelay--
		}
		keys := inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range keys {
			if k == ebiten.KeyR {
				continue
			}
			if s.pressDelay == 0 {
				s.transport.SendCommand(&transport.Command{
					Type:    CmdAction,
					Payload: ActionPayload{Key: k.String()},
				})
				s.pressDelay = config.Global().PressDelay
			}
		}
	}

	return nil
}

// Draw renders the level viewport and HUD.
func (s *SPClientState) Draw(screen *ebiten.Image) {
	cfg := config.Global()
	// Snap to nearest integer so every source pixel maps to the same
	// number of screen pixels — non-integer scales make pixel art wiggle.
	scale := math.Round(cfg.RenderScale)
	if scale < 1 {
		scale = 1
	}

	s.sim.Mu.RLock()
	if s.sim.Player != nil {
		pc := s.sim.Player.GetComponent("Position").(*component.PositionComponent)
		s.CameraX = pc.GetX()
		s.CameraY = pc.GetY()
	}
	tilesW := int(math.Ceil(float64(cfg.WorldWidth) / (float64(cfg.SpriteSizeW) * scale)))
	tilesH := int(math.Ceil(float64(cfg.WorldHeight) / (float64(cfg.SpriteSizeH) * scale)))
	levelImage := s.sim.Level.Render(
		s.CameraX, s.CameraY, s.sim.CurrentZ,
		tilesW, tilesH,
		false, true,
	)
	s.sim.Mu.RUnlock()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	screen.DrawImage(levelImage, op)

	s.gui.Draw(screen, s)
	s.inventoryView.Draw(screen)
}
