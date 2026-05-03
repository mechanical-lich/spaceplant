package game

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath"
	"github.com/mechanical-lich/spaceplant/internal/keybindings"
)

// TargetingCursor holds state for the targeting cursor mode.
type TargetingCursor struct {
	Active  bool
	Burst   bool
	enemies []*ecs.Entity
	idx     int
}

func newTargetingCursor() *TargetingCursor { return &TargetingCursor{} }

func (t *TargetingCursor) target() *ecs.Entity {
	if !t.Active || len(t.enemies) == 0 || t.idx < 0 || t.idx >= len(t.enemies) {
		return nil
	}
	return t.enemies[t.idx]
}

func (t *TargetingCursor) cycleNext() {
	if len(t.enemies) == 0 {
		return
	}
	t.idx = (t.idx + 1) % len(t.enemies)
}

func (t *TargetingCursor) exit() {
	t.Active = false
	t.Burst = false
	t.enemies = nil
	t.idx = 0
}

// enterTargetingMode switches to targeting mode, snapping the cursor to the
// nearest visible enemy in the player's facing direction, or the nearest overall.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) enterTargetingMode(burst bool) {
	if !playerHasRangedWeapon(s.sim.Player) {
		message.AddMessage("No ranged weapon equipped.")
		return
	}
	enemies := s.visibleEnemies()
	if len(enemies) == 0 {
		message.AddMessage("No targets in sight.")
		return
	}
	s.targeting.Active = true
	s.targeting.Burst = burst
	s.targeting.enemies = enemies
	s.targeting.idx = s.nearestEnemyInFacingDir(enemies)
}

// enterTargetingModeAtTile enters targeting mode with the cursor on the enemy
// at (tx, ty) if one exists, otherwise the nearest enemy in that direction.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) enterTargetingModeAtTile(tx, ty int) {
	if !playerHasRangedWeapon(s.sim.Player) {
		return
	}
	enemies := s.visibleEnemies()
	if len(enemies) == 0 {
		return
	}
	idx := -1
	for i, e := range enemies {
		epc := e.GetComponent(component.Position).(*component.PositionComponent)
		if epc.GetX() == tx && epc.GetY() == ty {
			idx = i
			break
		}
	}
	if idx < 0 && s.sim.Player != nil {
		pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
		px, py := pc.GetX(), pc.GetY()
		fdx := signInt(tx - px)
		fdy := signInt(ty - py)
		for i, e := range enemies {
			epc := e.GetComponent(component.Position).(*component.PositionComponent)
			if signInt(epc.GetX()-px) == fdx && signInt(epc.GetY()-py) == fdy {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		idx = 0
	}
	s.targeting.Active = true
	s.targeting.Burst = false
	s.targeting.enemies = enemies
	s.targeting.idx = idx
}

// updateTargeting handles input and refreshes state when targeting mode is active.
// Returns true if targeting mode is active (caller should skip normal input handling).
func (s *SPClientState) updateTargeting() bool {
	if !s.targeting.Active {
		return false
	}

	s.sim.Mu.RLock()
	s.targeting.enemies = s.visibleEnemies()
	if len(s.targeting.enemies) == 0 {
		s.targeting.exit()
		s.aimLineTiles = s.aimLineTiles[:0]
		s.sim.Mu.RUnlock()
		message.AddMessage("Target lost.")
		return true
	}
	if s.targeting.idx >= len(s.targeting.enemies) {
		s.targeting.idx = len(s.targeting.enemies) - 1
	}
	s.sim.Mu.RUnlock()

	kb := keybindings.Global()

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		s.targeting.cycleNext()
		return true
	}
if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || kb.IsJustPressed("fire") || kb.IsJustPressed("burst_fire") {
		s.confirmTargeting()
		return true
	}
	if kb.IsJustPressed("aimed_shot") {
		s.confirmTargetingAimed()
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		tx, ty, onWorld := s.mouseWorldTile()
		if onWorld {
			s.sim.Mu.RLock()
			s.selectTargetAtTile(tx, ty)
			s.sim.Mu.RUnlock()
		}
	}
	return true
}

// confirmTargeting fires at the current target and exits targeting mode.
func (s *SPClientState) confirmTargeting() {
	t := s.targeting.target()
	burst := s.targeting.Burst
	s.targeting.exit()
	s.aimLineTiles = s.aimLineTiles[:0]
	if t == nil {
		return
	}
	s.sim.Mu.RLock()
	epc := t.GetComponent(component.Position).(*component.PositionComponent)
	tx, ty := epc.GetX(), epc.GetY()
	s.sim.Mu.RUnlock()
	s.transport.SendCommand(&transport.Command{
		Type:    CmdMouseShoot,
		Payload: MouseShootPayload{TargetX: tx, TargetY: ty, Aimed: !burst, Burst: burst},
	})
	s.pressDelay = config.Global().PressDelay
}

// confirmTargetingAimed opens the body part picker for the current target.
func (s *SPClientState) confirmTargetingAimed() {
	t := s.targeting.target()
	if t == nil {
		return
	}
	s.sim.Mu.RLock()
	epc := t.GetComponent(component.Position).(*component.PositionComponent)
	tx, ty := epc.GetX(), epc.GetY()
	s.sim.Mu.RUnlock()
	s.targeting.exit()
	s.aimLineTiles = s.aimLineTiles[:0]
	s.aimedShotView.OnSelect = func(bodyPart string) {
		s.transport.SendCommand(&transport.Command{
			Type: CmdMouseShoot,
			Payload: MouseShootPayload{
				TargetX: tx, TargetY: ty,
				Aimed: true, AimedBodyPart: bodyPart,
			},
		})
		s.pressDelay = config.Global().PressDelay
	}
	s.aimedShotView.Open(t)
}

// selectTargetAtTile moves the cursor to the enemy at (tx, ty) if one exists.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) selectTargetAtTile(tx, ty int) {
	for i, e := range s.targeting.enemies {
		epc := e.GetComponent(component.Position).(*component.PositionComponent)
		if epc.GetX() == tx && epc.GetY() == ty {
			s.targeting.idx = i
			return
		}
	}
}

// visibleEnemies returns all living enemies visible to the player, sorted by distance.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) visibleEnemies() []*ecs.Entity {
	if s.sim.Player == nil {
		return nil
	}
	pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()
	var result []*ecs.Entity
	for _, e := range s.sim.Level.Level.GetEntities() {
		if e == nil || e == s.sim.Player {
			continue
		}
		if !e.HasComponent(component.Position) {
			continue
		}
		epc := e.GetComponent(component.Position).(*component.PositionComponent)
		if epc.GetZ() != pz {
			continue
		}
		if e.HasComponent(component.Dead) {
			continue
		}
		if !e.HasComponent(component.Body) {
			continue
		}
		ex, ey := epc.GetX(), epc.GetY()
		if !rlfov.Los(s.sim.Level.Level, px, py, ex, ey, pz) {
			continue
		}
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		epi := result[i].GetComponent(component.Position).(*component.PositionComponent)
		epj := result[j].GetComponent(component.Position).(*component.PositionComponent)
		di := absInt(epi.GetX()-px) + absInt(epi.GetY()-py)
		dj := absInt(epj.GetX()-px) + absInt(epj.GetY()-py)
		return di < dj
	})
	return result
}

// nearestEnemyInFacingDir returns the index of the nearest enemy in the player's
// facing direction, falling back to index 0 (globally nearest) if none are in front.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) nearestEnemyInFacingDir(enemies []*ecs.Entity) int {
	if s.sim.Player == nil || len(enemies) == 0 {
		return 0
	}
	pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	px, py := pc.GetX(), pc.GetY()
	fdx, fdy := 0, -1
	if s.sim.Player.HasComponent(component.Direction) {
		dc := s.sim.Player.GetComponent(component.Direction).(*component.DirectionComponent)
		switch dc.Direction {
		case 0:
			fdx, fdy = 1, 0
		case 1:
			fdx, fdy = 0, 1
		case 2:
			fdx, fdy = 0, -1
		case 3:
			fdx, fdy = -1, 0
		}
	}
	for i, e := range enemies {
		epc := e.GetComponent(component.Position).(*component.PositionComponent)
		ex, ey := epc.GetX(), epc.GetY()
		if signInt(ex-px) == fdx && signInt(ey-py) == fdy {
			return i
		}
	}
	return 0
}

// updateAimLineToTarget computes aimLineTiles along the Bresenham line to the current target.
// Must be called with sim.Mu.RLock held.
func (s *SPClientState) updateAimLineToTarget() {
	s.aimLineTiles = s.aimLineTiles[:0]
	t := s.targeting.target()
	if t == nil || s.sim.Player == nil {
		return
	}
	pc := s.sim.Player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()
	epc := t.GetComponent(component.Position).(*component.PositionComponent)
	tx, ty := epc.GetX(), epc.GetY()

	const maxRange = 16
	for _, tile := range rlmath.BresenhamLine(px, py, tx, ty) {
		wx, wy := tile[0], tile[1]
		s.aimLineTiles = append(s.aimLineTiles, [2]int{wx, wy})
		if len(s.aimLineTiles) >= maxRange {
			break
		}
		if s.sim.Level.IsTileSolid(wx, wy, pz) {
			break
		}
		if e := s.sim.Level.Level.GetSolidEntityAt(wx, wy, pz); e != nil && e != s.sim.Player {
			if e.HasComponent(component.Door) {
				dc := e.GetComponent(component.Door).(*component.DoorComponent)
				if dc.Open {
					continue
				}
			}
			break
		}
	}
}

// drawTargetCursor draws a yellow reticle on the current target's tile.
func (s *SPClientState) drawTargetCursor(screen *ebiten.Image) {
	t := s.targeting.target()
	if t == nil {
		return
	}
	s.sim.Mu.RLock()
	epc := t.GetComponent(component.Position).(*component.PositionComponent)
	tile := [2]int{epc.GetX(), epc.GetY()}
	s.sim.Mu.RUnlock()
	s.drawReticleTiles(screen, [][2]int{tile}, 1, 1, 0, 1, 1, 1, 0, 1)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
