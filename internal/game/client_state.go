package game

import (
	"fmt"

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
	tick          int
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

	s.gui.Update(s)
	s.inventoryView.Update()
	s.tick++

	fps := ebiten.ActualFPS()
	tps := ebiten.ActualTPS()
	title := fmt.Sprintf("%s - FPS: %.1f TPS: %.1f", "Space Plants!", fps, tps)
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

	// Sprite animation (every 20 frames).
	if s.tick%20 == 0 {
		for _, entity := range s.sim.Level.Entities {
			if entity.HasComponent("AppearanceComponent") {
				ac := entity.GetComponent("AppearanceComponent").(*component.AppearanceComponent)
				ac.Update()
			}
		}
	}

	// Send movement/action commands only when it is the player's turn.
	if s.sim.Player != nil && s.sim.Player.HasComponent("MyTurn") {
		if s.pressDelay > 0 {
			s.pressDelay--
		}
		keys := inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range keys {
			if s.pressDelay == 0 {
				s.transport.SendCommand(&transport.Command{
					Type:    CmdAction,
					Payload: ActionPayload{Key: k.String()},
				})
				s.pressDelay = config.PressDelay
			}
		}
	}

	return nil
}

// Draw renders the level viewport and HUD.
func (s *SPClientState) Draw(screen *ebiten.Image) {
	s.sim.Mu.RLock()
	if s.sim.Player != nil {
		pc := s.sim.Player.GetComponent("Position").(*component.PositionComponent)
		s.CameraX = pc.GetX()
		s.CameraY = pc.GetY()
	}
	levelImage := s.sim.Level.Render(
		s.CameraX, s.CameraY, s.sim.CurrentZ,
		config.GameWidth/config.SpriteWidth, config.GameHeight/config.SpriteHeight,
		false, true,
	)
	s.sim.Mu.RUnlock()

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(levelImage, op)

	s.gui.Draw(screen, s)
	s.inventoryView.Draw(screen)
}
