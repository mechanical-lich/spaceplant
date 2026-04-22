package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/client"
	"github.com/mechanical-lich/mlge/ecs"
	mlgeevent "github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/eventsystem"
	"github.com/mechanical-lich/spaceplant/internal/keybindings"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
)

// compile-time assertion
var _ client.ClientState = (*SPClientState)(nil)

// SPClientState is the graphical (Ebiten) client state.
// It polls keyboard input and forwards action commands to the server,
// then renders the level and HUD each frame.
type SPClientState struct {
	sim              *SimWorld
	simState         *MainSimState
	transport        transport.ClientTransport
	mainView         *GUIViewMain
	classView        *ClassUpgradeView
	statsView        *CharacterStatsView
	reloadView       *ReloadView
	aimedShotView    *AimedShotView
	nearbyLootView   *NearbyLootView
	characterCreator *CharacterCreator
	deathModal       *DeathModal
	winModal         *WinModal
	pauseMenu        *PauseMenu
	cheatModal       *CheatModal
	CameraX          int
	CameraY          int
	pressDelay       int
	returnToTitle    bool
	lastSavedTurn    int
}

// NewSPClientState creates a ready-to-use graphical client state.
// If the player hasn't been spawned yet, the character creator is shown first.
func NewSPClientState(sim *SimWorld, simState *MainSimState, t transport.ClientTransport) *SPClientState {
	cs := &SPClientState{
		sim:       sim,
		simState:  simState,
		transport: t,
		mainView:  &GUIViewMain{},
	}

	if sim.Player == nil {
		cs.characterCreator = NewCharacterCreator()
		cs.characterCreator.OnComplete = func(data CharacterData) {
			if err := sim.SpawnPlayer(data); err != nil {
				fmt.Printf("Error spawning player: %v\n", err)
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
	s.aimedShotView = NewAimedShotView()
	s.aimedShotView.OnSelect = func(bodyPart string) {
		s.transport.SendCommand(&transport.Command{
			Type:    CmdAimedShot,
			Payload: AimedShotPayload{BodyPart: bodyPart},
		})
	}
	s.nearbyLootView = NewNearbyLootView()
	s.nearbyLootView.OnPickup = func(item *ecs.Entity, tx, ty, tz int) {
		s.transport.SendCommand(&transport.Command{
			Type:    CmdPickupItem,
			Payload: PickupItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
		})
	}
	s.nearbyLootView.OnEquip = func(item *ecs.Entity, tx, ty, tz int) {
		s.transport.SendCommand(&transport.Command{
			Type:    CmdEquipItem,
			Payload: EquipItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
		})
	}
	pc := s.sim.Player.GetComponent("Position").(*component.PositionComponent)
	s.CameraX = pc.GetX()
	s.CameraY = pc.GetY()

	s.deathModal = newDeathModal()
	s.deathModal.OnReturnToTitle = func() {
		// Convert the player to a persistent corpse on the station, then save.
		if s.sim.PlayerRunID != "" {
			s.sim.ConvertPlayerToCorpse()
			if err := SaveStation(s.sim, "saves"); err != nil {
				fmt.Printf("Save station after death failed: %v\n", err)
			}
			if err := GraveyardPlayerRun("saves", s.sim.PlayerRunID); err != nil {
				fmt.Printf("Graveyard player run failed: %v\n", err)
			}
		}
		s.returnToTitle = true
	}
	eventsystem.EventManager.RegisterListener(
		&gameLostListener{modal: s.deathModal},
		eventsystem.GameLost,
	)

	s.winModal = newWinModal()
	s.winModal.OnReturnToTitle = func() {
		if s.sim.PlayerRunID != "" {
			if err := SaveStation(s.sim, "saves"); err != nil {
				fmt.Printf("Save station after win failed: %v\n", err)
			}
			if err := GraveyardWonPlayerRun("saves", s.sim.PlayerRunID, s.winModal.Outcome); err != nil {
				fmt.Printf("Graveyard won player run failed: %v\n", err)
			}
		}
		s.returnToTitle = true
	}
	eventsystem.EventManager.RegisterListener(
		&gameWonListener{modal: s.winModal},
		eventsystem.GameWon,
	)

	s.cheatModal = newCheatModal(s.sim)
	s.pauseMenu = newPauseMenu()
	s.pauseMenu.OnSave = func() {
		if err := SaveAll(s.sim, "saves"); err != nil {
			fmt.Printf("Save failed: %v\n", err)
		} else {
			message.AddMessage("Game saved.")
		}
	}
	s.pauseMenu.OnReturnToTitle = func() {
		s.returnToTitle = true
	}
}

func (s *SPClientState) Done() bool { return s.returnToTitle }

// playerDeathMessage returns a short cause-of-death string for the death modal.
// Must be called with at least s.sim.Mu.RLock held.
func (s *SPClientState) playerDeathMessage() string {
	player := s.sim.Player
	if player == nil {
		return "You have died."
	}
	if player.HasComponent("Body") {
		bc := player.GetComponent("Body").(*rlcomponents.BodyComponent)
		for _, part := range bc.Parts {
			if part.Broken && part.KillsWhenBroken {
				return "Your " + part.Name + " was destroyed."
			}
			if part.Amputated && part.KillsWhenAmputated {
				return "Your " + part.Name + " was amputated."
			}
		}
	}
	return "You have died."
}

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

	// If "Return to Title" was clicked in the death modal, push the title screen.
	if s.returnToTitle {
		return NewTitleScreenState(s.sim, s.simState, s.transport)
	}

	fps := ebiten.ActualFPS()
	tps := ebiten.ActualTPS()
	s.sim.Mu.RLock()
	tickCount := s.sim.TickCount
	turnCount := s.sim.TurnCount
	s.sim.Mu.RUnlock()
	title := fmt.Sprintf("%s - Turn: %d Tick: %d FPS: %.1f TPS: %.1f", "Space Plants!", turnCount, tickCount, fps, tps)
	ebiten.SetWindowTitle(title)

	// Auto-save every 5 player turns.
	if turnCount > 0 && turnCount != s.lastSavedTurn && turnCount%5 == 0 {
		s.lastSavedTurn = turnCount
		go func() {
			if err := SaveAll(s.sim, "saves"); err != nil {
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		}()
	}

	shift := ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	// Shift+ESC: open developer cheat console.
	if shift && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if !s.cheatModal.Visible {
			s.cheatModal.Open()
		}
		return nil
	}

	// ESC: close innermost open modal, or open the pause menu.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.nearbyLootView.Visible {
			s.nearbyLootView.Visible = false
			return nil
		}
		if s.aimedShotView.Visible {
			s.aimedShotView.Visible = false
			return nil
		}
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
		if s.pauseMenu.Visible {
			s.pauseMenu.Visible = false
			return nil
		}
		// Nothing else open — open the pause menu.
		s.pauseMenu.Open()
		return nil
	}

	// Inventory actions (work outside hasTurn).
	kb := keybindings.Global()
	if !s.cheatModal.Visible {
		if kb.IsJustPressed("inventory") {
			if !s.classView.Visible && !s.statsView.Visible {
				s.statsView.SetNearbyEntity(s.nearbyInventoryEntity())
				s.statsView.OpenToInventory()
				return nil
			}
		}
		if kb.IsJustPressed("character_overview") {
			if !s.classView.Visible && !s.statsView.Visible {
				s.statsView.Open()
				return nil
			}
		}
	}

	// GUI, inventory, class view, stats view, and turn check — RLock is sufficient.
	s.sim.Mu.RLock()
	s.mainView.Update(s)
	s.classView.Update()
	s.statsView.Update()
	s.reloadView.Update()
	s.aimedShotView.Update()
	s.nearbyLootView.Update()
	hasTurn := s.sim.Player != nil && s.sim.Player.HasComponent("MyTurn")
	// Fallback: catch deaths where no modal fired yet (e.g. race with server
	// goroutine, or edge cases with no DeathEvent). Evaluate conditions so JSON
	// rules are respected even on this path.
	if !s.deathModal.Visible && !s.winModal.Visible &&
		s.sim.Player != nil && s.sim.Player.HasComponent("Dead") {
		ctx := s.sim.BuildEvalContext()
		if rule, ok := wincondition.Active().EvalPlayerDeath(ctx); ok {
			wincondition.FireRule(rule, s.playerDeathMessage())
		}
	}
	s.sim.Mu.RUnlock()

	// Update death modal and pause menu outside the lock — their callbacks may
	// set returnToTitle. Check immediately after so Done() and Update() agree
	// on the same frame (avoiding an empty-stack black screen).
	s.deathModal.Update()
	s.winModal.Update()
	s.pauseMenu.Update()
	s.cheatModal.Update()
	if s.returnToTitle {
		return NewTitleScreenState(s.sim, s.simState, s.transport)
	}

	// Block game input while any modal is open.
	if s.classView.Visible || s.statsView.Visible || s.reloadView.Visible || s.aimedShotView.Visible || s.nearbyLootView.Visible || s.deathModal.Visible || s.winModal.Visible || s.pauseMenu.Visible || s.cheatModal.Visible {
		return nil
	}

	// Send movement/action commands only when it is the player's turn.
	if hasTurn {
		// Rush is a just-pressed toggle, not a repeating held action.
		if kb.IsJustPressed("rush") {
			s.transport.SendCommand(&transport.Command{
				Type:    CmdAction,
				Payload: ActionPayload{Key: "rush"},
			})
		}
		// Reload modal is also just-pressed.
		if kb.IsJustPressed("reload") {
			s.reloadView.Open()
			return nil
		}

		if s.pressDelay > 0 {
			s.pressDelay--
		}
		keys := inpututil.AppendPressedKeys([]ebiten.Key{})
		for _, k := range keys {
			if isModifierKey(k) {
				continue
			}
			if s.pressDelay == 0 {
				action := kb.ActionFor(k.String(), shift)
				switch action {
				case "rush", "reload", "inventory", "character_overview":
					// handled above or outside hasTurn — skip
				case "class_upgrade":
					s.classView.Open()
					s.pressDelay = config.Global().PressDelay
				case "pickup":
					s.sim.Mu.RLock()
					hasNearby := s.hasNearbyItems()
					s.sim.Mu.RUnlock()
					if hasNearby {
						s.sim.Mu.RLock()
						s.nearbyLootView.Open(s.sim.Player, s.sim.Level)
						s.sim.Mu.RUnlock()
					} else {
						s.transport.SendCommand(&transport.Command{
							Type:    CmdAction,
							Payload: ActionPayload{Key: "pickup"},
						})
					}
					s.pressDelay = config.Global().PressDelay
				case "aimed_shot":
					s.sim.Mu.RLock()
					target := s.rayTarget()
					s.sim.Mu.RUnlock()
					if target == nil {
						message.AddMessage("Nothing to aim at.")
					} else {
						s.aimedShotView.Open(target)
					}
					s.pressDelay = config.Global().PressDelay
				case "":
					// No binding — pass raw key string through for skill hotkeys.
					keyStr := k.String()
					if len(keyStr) == 1 && keyStr[0] >= 'A' && keyStr[0] <= 'Z' && !shift {
						keyStr = strings.ToLower(keyStr)
					}
					s.transport.SendCommand(&transport.Command{
						Type:    CmdAction,
						Payload: ActionPayload{Key: keyStr},
					})
					s.pressDelay = config.Global().PressDelay
				default:
					s.transport.SendCommand(&transport.Command{
						Type:    CmdAction,
						Payload: ActionPayload{Key: action},
					})
					s.pressDelay = config.Global().PressDelay
				}
			}
		}
	}

	return nil
}

// nearbyInventoryEntity returns the best inventory-bearing entity to show in
// the nearby panel:
//   - Dead entities (corpses) must be on the player's exact tile — the player
//     can only stand on them once they're dead, which acts as a loot gate.
//   - Inanimate containers (lockers, crates, etc.) are found within 1 tile so
//     the player doesn't have to stand inside them.
func (s *SPClientState) nearbyInventoryEntity() *ecs.Entity {
	if s.sim.Player == nil {
		return nil
	}
	pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()

	hasInv := func(e *ecs.Entity) bool {
		return e.HasComponent(component.BodyInventory) || e.HasComponent(component.Inventory)
	}

	// First pass: same tile only — picks up dead entities (and same-tile containers).
	var buf []*ecs.Entity
	s.sim.Level.Level.GetEntitiesAt(px, py, pz, &buf)
	for _, e := range buf {
		if e != s.sim.Player && hasInv(e) {
			return e
		}
	}

	// Second pass: adjacent tiles — inanimate containers only (not live creatures).
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			buf = buf[:0]
			s.sim.Level.Level.GetEntitiesAt(px+dx, py+dy, pz, &buf)
			for _, e := range buf {
				if e != s.sim.Player && hasInv(e) && e.HasComponent(component.Inanimate) {
					return e
				}
			}
		}
	}

	return nil
}

// hasNearbyItems reports whether any item entities exist on the player's tile
// or any of the 8 adjacent tiles. Must be called with at least RLock held.
func (s *SPClientState) hasNearbyItems() bool {
	player := s.sim.Player
	if player == nil {
		return false
	}
	pc := player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			var buf []*ecs.Entity
			s.sim.Level.Level.GetEntitiesAt(px+dx, py+dy, pz, &buf)
			for _, e := range buf {
				if e != player && e.HasComponent(component.Item) {
					return true
				}
			}
		}
	}
	return false
}

// rayTarget walks the facing direction from the player and returns the first
// solid entity encountered within weapon range, or nil if none is found.
// Must be called with at least s.sim.Mu.RLock held.
func (s *SPClientState) rayTarget() *ecs.Entity {
	player := s.sim.Player
	if player == nil {
		return nil
	}
	pc := player.GetComponent(component.Position).(*component.PositionComponent)
	x, y, z := pc.GetX(), pc.GetY(), pc.GetZ()

	dx, dy := 0, -1 // default: up
	if player.HasComponent(component.Direction) {
		dc := player.GetComponent(component.Direction).(*component.DirectionComponent)
		switch dc.Direction {
		case 0:
			dx, dy = 1, 0
		case 1:
			dx, dy = 0, 1
		case 2:
			dx, dy = 0, -1
		case 3:
			dx, dy = -1, 0
		}
	}

	const maxRange = 16
	for i := 1; i <= maxRange; i++ {
		tx, ty := x+dx*i, y+dy*i
		if s.sim.Level.IsTileSolid(tx, ty, z) {
			break
		}
		if e := s.sim.Level.Level.GetSolidEntityAt(tx, ty, z); e != nil && e != player {
			if e.HasComponent(component.Door) {
				dc := e.GetComponent(component.Door).(*component.DoorComponent)
				if dc.Open {
					continue
				}
			}
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
	tilesW := int(math.Ceil(float64(cfg.WorldWidth) / (float64(cfg.TileSizeW) * scale)))
	tilesH := int(math.Ceil(float64(cfg.WorldHeight) / (float64(cfg.TileSizeH) * scale)))
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
	s.aimedShotView.Draw(screen)
	s.nearbyLootView.Draw(screen)
	s.deathModal.Draw(screen)
	s.winModal.Draw(screen)
	s.pauseMenu.Draw(screen)
	s.cheatModal.Draw(screen)
}
