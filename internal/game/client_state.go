package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/client"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// compile-time assertion
var _ client.ClientState = (*SPClientState)(nil)

// SPClientState is the graphical (Ebiten) client state.
// It polls keyboard input and forwards action commands to the server,
// then renders the level and HUD each frame.
type SPClientState struct {
	sim              *SimWorld
	transport        transport.ClientTransport
	mainView         *GUIViewMain
	classView        *ClassUpgradeView
	statsView        *CharacterStatsView
	reloadView       *ReloadView
	characterCreator *CharacterCreator
	CameraX          int
	CameraY          int
	pressDelay       int
}

// NewSPClientState creates a ready-to-use graphical client state.
// If the player hasn't been spawned yet, the character creator is shown first.
func NewSPClientState(sim *SimWorld, t transport.ClientTransport) *SPClientState {
	cs := &SPClientState{
		sim:       sim,
		transport: t,
		mainView:  &GUIViewMain{},
	}

	if sim.Player == nil {
		cs.characterCreator = NewCharacterCreator()
		cs.characterCreator.OnComplete = func(data CharacterData) {
			if err := sim.SpawnPlayer(data); err != nil {
				return
			}
			cs.initGameViews()
		}
	} else {
		cs.initGameViews()
	}

	return cs
}

// initGameViews creates the class and stats views once the player exists.
func (s *SPClientState) initGameViews() {
	s.classView = NewClassUpgradeView(s.sim.Player)
	s.statsView = NewCharacterStatsView(s.sim.Player)
	s.reloadView = NewReloadView(s.sim.Player)
	s.reloadView.OnReload = func(weaponItem, ammoItem *ecs.Entity) {
		s.transport.SendCommand(&transport.Command{
			Type:    CmdReload,
			Payload: ReloadPayload{WeaponItem: weaponItem, AmmoItem: ammoItem},
		})
	}
	pc := s.sim.Player.GetComponent("Position").(*component.PositionComponent)
	s.CameraX = pc.GetX()
	s.CameraY = pc.GetY()
}

func (s *SPClientState) Done() bool { return false }

func isModifierKey(k ebiten.Key) bool {
	switch k {
	case ebiten.KeyShift, ebiten.KeyShiftLeft, ebiten.KeyShiftRight,
		ebiten.KeyControl, ebiten.KeyControlLeft, ebiten.KeyControlRight,
		ebiten.KeyAlt, ebiten.KeyAltLeft, ebiten.KeyAltRight,
		ebiten.KeyMeta, ebiten.KeyMetaLeft, ebiten.KeyMetaRight:
		return true
	}
	return false
}

// Update is called every Ebiten frame. It handles input, sends commands to the
// server, animates sprites, and updates the HUD.
func (s *SPClientState) Update(_ *transport.Snapshot) client.ClientState {
	mlgeevent.GetQueuedInstance().HandleQueue()

	// Show character creator until the player is spawned.
	if s.sim.Player == nil {
		s.characterCreator.Update()
		return nil
	}

	fps := ebiten.ActualFPS()
	tps := ebiten.ActualTPS()
	s.sim.Mu.RLock()
	tickCount := s.sim.TickCount
	turnCount := s.sim.TurnCount
	s.sim.Mu.RUnlock()
	title := fmt.Sprintf("%s - Turn: %d Tick: %d FPS: %.1f TPS: %.1f", "Space Plants!", turnCount, tickCount, fps, tps)
	ebiten.SetWindowTitle(title)

	shift := ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	// Close modals on Escape (innermost first).
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.reloadView.Visible {
			s.reloadView.Visible = false
			return nil
		}
		if s.statsView.Visible {
			s.statsView.Visible = false
			return nil
		}
		if s.classView.Visible {
			s.classView.Visible = false
			return nil
		}
	}

	// I opens the stats modal (inventory tab); Shift+I opens to the overview tab.
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		if !s.classView.Visible && !s.statsView.Visible {
			if shift {
				s.statsView.Open()
			} else {
				s.statsView.SetNearbyEntity(s.nearbyInventoryEntity())
				s.statsView.OpenToInventory()
			}
			return nil
		}
	}

	// GUI, inventory, class view, stats view, and turn check — RLock is sufficient.
	s.sim.Mu.RLock()
	s.mainView.Update(s)
	s.classView.Update()
	s.statsView.Update()
	s.reloadView.Update()
	hasTurn := s.sim.Player != nil && s.sim.Player.HasComponent("MyTurn")
	s.sim.Mu.RUnlock()

	// Block game input while any modal is open.
	if s.classView.Visible || s.statsView.Visible || s.reloadView.Visible {
		return nil
	}

	// Send movement/action commands only when it is the player's turn.
	if hasTurn {
		// Shift+R opens the reload modal; unmodified R toggles rush.
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			if shift {
				s.reloadView.Open()
				return nil
			}
			s.transport.SendCommand(&transport.Command{
				Type:    CmdAction,
				Payload: ActionPayload{Key: "r"},
			})
		}

		if s.pressDelay > 0 {
			s.pressDelay--
		}
		keys := inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range keys {
			if k == ebiten.KeyR || isModifierKey(k) {
				continue
			}
			if s.pressDelay == 0 {
				keyStr := k.String()
				// Single letter keys: lowercase without shift, uppercase with shift.
				if len(keyStr) == 1 && keyStr[0] >= 'A' && keyStr[0] <= 'Z' {
					if !shift {
						keyStr = strings.ToLower(keyStr)
					}
				}
				// Shift+C opens the class upgrade modal.
				if keyStr == "C" {
					s.classView.Open()
					s.pressDelay = config.Global().PressDelay
					continue
				}
				s.transport.SendCommand(&transport.Command{
					Type:    CmdAction,
					Payload: ActionPayload{Key: keyStr},
				})
				s.pressDelay = config.Global().PressDelay
			}
		}
	}

	return nil
}

// nearbyInventoryEntity returns the first entity on the player's tile (other
// than the player) that has an inventory, or nil if none exists.
func (s *SPClientState) nearbyInventoryEntity() *ecs.Entity {
	if s.sim.Player == nil {
		return nil
	}
	pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	var buf []*ecs.Entity
	s.sim.Level.Level.GetEntitiesAt(pc.GetX(), pc.GetY(), pc.GetZ(), &buf)
	for _, e := range buf {
		if e == s.sim.Player {
			continue
		}
		if e.HasComponent(component.BodyInventory) || e.HasComponent(component.Inventory) {
			return e
		}
	}
	return nil
}

// Draw renders the level viewport and HUD.
func (s *SPClientState) Draw(screen *ebiten.Image) {
	// Show character creator until the player is spawned.
	if s.sim.Player == nil {
		s.characterCreator.Draw(screen)
		return
	}

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

	s.mainView.Draw(screen, s)
	s.classView.Draw(screen)
	s.statsView.Draw(screen)
	s.reloadView.Draw(screen)
}
