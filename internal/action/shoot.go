package action

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/energy"
	"github.com/mechanical-lich/spaceplant/internal/entityhelpers"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ShootAction fires a ranged weapon in the direction the entity is currently facing.
//
//	Aimed=true  — deliberate aimed shot: higher energy cost, CS bonus (+10).
//	Burst=true  — burst fire: one CS roll per round with increasing recoil penalty (-5 per round after first).
//
// Burst and Aimed are mutually exclusive; Burst takes priority if both are set.
// SpreadAngle on the equipped weapon fires additional diagonal lines (1 = 3-wide, 2 = 5-wide).
type ShootAction struct {
	Aimed         bool
	Burst         bool
	AimedBodyPart string // non-empty: targeted aimed shot at a specific body part
	DX, DY        int    // non-zero: override facing direction (legacy keyboard fire)
	TargetX, TargetY int // world tile of locked target; when set, Bresenham line is used
	HasTarget     bool   // true when TargetX/TargetY are valid
}

// Range-band CS modifiers.
const (
	csPointBlankPenalty = -20
	csLongRangePenalty  = -15
	csAimedShotBonus    = +10
	csBurstBonus        = +15 // base bonus on first round of a burst
	csBurstRecoil       = -5  // recoil penalty per subsequent round
)

// spreadPenMult is the Pen multiplier applied to non-centre spread lines (60%).
const spreadPenMult = 0.6

func (a ShootAction) Cost(_ *ecs.Entity, _ *world.Level) int {
	if a.Burst {
		return energy.CostBurst
	}
	if a.AimedBodyPart != "" {
		return energy.CostAimedTargeted
	}
	if a.Aimed {
		return energy.CostAimed
	}
	return energy.CostAttack
}

func (a ShootAction) Available(entity *ecs.Entity, _ *world.Level) bool {
	wc := equippedRangedWeapon(entity)
	if wc == nil {
		return false
	}
	if a.Burst {
		return wc.BurstSize >= 2
	}
	return true
}

func (a ShootAction) Execute(entity *ecs.Entity, level *world.Level) error {
	cost := energy.CostAttack
	if a.Burst {
		cost = energy.CostBurst
	} else if a.Aimed {
		cost = energy.CostAimed
	}

	wc := equippedRangedWeapon(entity)
	if wc == nil {
		message.AddMessage("You don't have a ranged weapon equipped.")
		entity.RemoveComponent(rlcomponents.TurnTaken)
		return nil
	}

	dx, dy := facingDeltas(entity)
	if a.DX != 0 || a.DY != 0 {
		dx, dy = a.DX, a.DY
	}

	// Bresenham line overrides direction-based walk when a target tile is locked.
	var line [][2]int
	if a.HasTarget {
		pc := entity.GetComponent(component.Position).(*component.PositionComponent)
		line = rlmath.BresenhamLine(pc.GetX(), pc.GetY(), a.TargetX, a.TargetY)
		// Derive dx/dy from the target vector for spread calculations.
		vx := a.TargetX - pc.GetX()
		vy := a.TargetY - pc.GetY()
		dx, dy = signInt(vx), signInt(vy)
	}

	if !a.HasTarget && dx == 0 && dy == 0 {
		message.AddMessage("No direction to fire.")
		rlenergy.SetActionCost(entity, energy.CostQuick)
		return nil
	}

	maxRange := wc.Range
	if maxRange <= 0 {
		maxRange = 8
	}

	switch {
	case a.Burst:
		if wc.MaxMagazine > 0 && wc.Magazine <= 0 {
			message.AddMessage("*click* Out of ammo. Press Shift+R to reload.")
			rlenergy.SetActionCost(entity, energy.CostQuick)
			return nil
		}
		execBurst(entity, level, wc, dx, dy, maxRange, line)
	default:
		// Single shot (snap or aimed).
		if wc.MaxMagazine > 0 && wc.Magazine <= 0 {
			message.AddMessage("*click* Out of ammo. Press Shift+R to reload.")
			rlenergy.SetActionCost(entity, energy.CostQuick)
			return nil
		}
		csBonus := 0
		if a.Aimed || a.AimedBodyPart != "" {
			csBonus += csAimedShotBonus
			if entity.HasComponent(component.Skill) {
				sc := entity.GetComponent(component.Skill).(*component.SkillComponent)
				if sc.HasSkill("deadshot") {
					csBonus += csAimedShotBonus // extra +10 for deadshot
				}
			}
		}
		execAuto(entity, level, wc, dx, dy, maxRange, csBonus, a.AimedBodyPart, line)
	}

	rlenergy.SetActionCost(entity, cost)
	return nil
}

// execAuto fires wc.AutoRounds rounds per trigger pull (defaults to 1).
// Unlike burst, auto fire is the normal shot — no recoil stacking, no player choice.
// line is the Bresenham tile path when targeting mode was used; nil falls back to direction.
func execAuto(entity *ecs.Entity, level *world.Level, wc *component.WeaponComponent, dx, dy, maxRange, csBonus int, aimedBodyPart string, line [][2]int) {
	rounds := wc.AutoRounds
	if rounds < 2 {
		execSpread(entity, level, wc, dx, dy, maxRange, csBonus, 1.0, aimedBodyPart, line)
		if wc.MaxMagazine > 0 {
			wc.Magazine--
		}
		return
	}
	for i := 0; i < rounds; i++ {
		if wc.MaxMagazine > 0 && wc.Magazine <= 0 {
			message.AddMessage("Weapon empty.")
			return
		}
		execSpread(entity, level, wc, dx, dy, maxRange, csBonus, 1.0, aimedBodyPart, line)
		if wc.MaxMagazine > 0 {
			wc.Magazine--
		}
	}
}

// execBurst fires wc.BurstSize rounds along the facing line.
// Each round gets the burst CS bonus minus accumulated recoil.
// Rounds after the first may walk to the next target if the primary is dead.
func execBurst(entity *ecs.Entity, level *world.Level, wc *component.WeaponComponent, dx, dy, maxRange int, line [][2]int) {
	if wc.BurstSize < 2 {
		message.AddMessage("Weapon does not support burst fire.")
		return
	}

	rounds := wc.BurstSize
	for i := 0; i < rounds; i++ {
		if wc.MaxMagazine > 0 && wc.Magazine <= 0 {
			message.AddMessage("Weapon empty mid-burst.")
			return
		}
		csBonus := csBurstBonus + csBurstRecoil*i
		hit := fireLineAt(entity, level, wc, dx, dy, maxRange, csBonus, 1.0, "", line)
		if wc.MaxMagazine > 0 {
			wc.Magazine--
		}
		if !hit && i == 0 {
			// First round missed into the void — no point continuing.
			return
		}
	}
}

// execSpread fires the primary line and, if wc.SpreadAngle > 0, additional
// diagonal spread lines. Spread lines deal reduced Pen.
// The Bresenham line (if any) applies only to the primary; spread lines always use direction.
func execSpread(entity *ecs.Entity, level *world.Level, wc *component.WeaponComponent, dx, dy, maxRange, csBonus int, penMult float64, aimedBodyPart string, line [][2]int) {
	// Only the primary line carries the aimed body part and Bresenham path.
	fireLineAt(entity, level, wc, dx, dy, maxRange, csBonus, penMult, aimedBodyPart, line)

	if wc.SpreadAngle <= 0 {
		return
	}

	// Perpendicular unit vector.
	px, py := dy, -dx

	for offset := 1; offset <= wc.SpreadAngle; offset++ {
		// Diagonal directions formed by combining facing + perpendicular offset.
		ldx := dx + px*offset
		ldy := dy + py*offset
		fireLineAt(entity, level, wc, ldx, ldy, maxRange, csBonus, spreadPenMult, "", nil)

		rdx := dx - px*offset
		rdy := dy - py*offset
		fireLineAt(entity, level, wc, rdx, rdy, maxRange, csBonus, spreadPenMult, "", nil)
	}
}

// fireLineAt walks the shot path and fires at the first hittable entity.
// When line is non-nil it walks that Bresenham tile sequence; otherwise it steps
// in direction (dx, dy) up to maxRange tiles.
// penMult scales Pen (1.0 for normal, <1.0 for spread pellets).
// Returns true if a target was found (whether or not the attack landed).
func fireLineAt(entity *ecs.Entity, level *world.Level, wc *component.WeaponComponent, dx, dy, maxRange, csBonus int, penMult float64, aimedBodyPart string, line [][2]int) bool {
	pc := entity.GetComponent(component.Position).(*component.PositionComponent)
	z := pc.GetZ()

	// Build the tile sequence to walk.
	type step struct{ x, y, dist int }
	var steps []step
	if len(line) > 0 {
		for i, t := range line {
			if i >= maxRange {
				break
			}
			steps = append(steps, step{t[0], t[1], i + 1})
		}
	} else {
		for i := 1; i <= maxRange; i++ {
			steps = append(steps, step{pc.GetX() + dx*i, pc.GetY() + dy*i, i})
		}
	}

	var target *ecs.Entity
	var targetDist int
	for _, s := range steps {
		tx, ty := s.x, s.y
		if level.IsTileSolid(tx, ty, z) {
			// Still flash the blocking tile so the tracer visibly hits it.
			addShotTrail(level, tx, ty, z)
			break
		}
		addShotTrail(level, tx, ty, z)
		candidate := level.Level.GetSolidEntityAt(tx, ty, z)
		if candidate != nil && candidate != entity {
			// Open doors don't block shots.
			if candidate.HasComponent(component.Door) {
				dc := candidate.GetComponent(component.Door).(*component.DoorComponent)
				if dc.Open {
					continue
				}
			}
			target = candidate
			targetDist = s.dist
			break
		}
	}

	if target == nil {
		message.AddMessage("You fire into the void.")
		return false
	}

	// Range-band modifier on top of caller-supplied csBonus.
	rb := rangeBandBonus(targetDist, maxRange, wc)

	// Apply Pen multiplier by temporarily adjusting — we need to pass an
	// effective pen to HitRanged. Build a shallow copy with scaled Pen.
	effectiveWC := *wc
	if penMult != 1.0 {
		effectiveWC.Penetration = int(float64(wc.Penetration) * penMult)
		if effectiveWC.Penetration < 1 {
			effectiveWC.Penetration = 1
		}
	}

	if aimedBodyPart != "" {
		entityhelpers.HitRangedTargeted(level, entity, target, &effectiveWC, csBonus+rb, aimedBodyPart)
	} else {
		entityhelpers.HitRanged(level, entity, target, &effectiveWC, csBonus+rb)
	}
	return true
}

// addShotTrail places a brief flash TileAnim on a tile to show the bullet path.
func addShotTrail(level *world.Level, x, y, z int) {
	level.AddTileAnim(x, y, z, &world.TileAnim{
		SpriteX:    384,
		SpriteY:    0,
		Resource:   "fx",
		FrameCount: 4,
		FrameSpeed: 1,
		TTL:        4,
	})
}

// rangeBandBonus returns the CS modifier for the given distance and weapon range.
// Uses the weapon's RangeBands if defined; falls back to global constants otherwise.
func rangeBandBonus(dist, maxRange int, wc *component.WeaponComponent) int {
	if mod, ok := wc.RangeBandCSMod(dist); ok {
		return mod
	}
	// Global fallback.
	if dist <= 1 {
		return csPointBlankPenalty
	}
	if dist > maxRange/2 {
		return csLongRangePenalty
	}
	return 0
}

// facingDeltas returns the dx,dy based on the entity's DirectionComponent.
// Direction values: 0=right, 1=down, 2=up, 3=left (from rlentity.Face).
func facingDeltas(entity *ecs.Entity) (int, int) {
	if !entity.HasComponent(component.Direction) {
		return 0, -1 // default: up
	}
	dc := entity.GetComponent(component.Direction).(*component.DirectionComponent)
	switch dc.Direction {
	case 0:
		return 1, 0 // right
	case 1:
		return 0, 1 // down
	case 2:
		return 0, -1 // up
	case 3:
		return -1, 0 // left
	default:
		return 0, -1
	}
}

func signInt(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// equippedMeleeWeapon returns the first equipped weapon with Ranged=false, or nil.
func equippedMeleeWeapon(entity *ecs.Entity) *component.WeaponComponent {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return wc
				}
			}
		}
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		for _, item := range []*ecs.Entity{inv.RightHand, inv.LeftHand} {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if !wc.Ranged {
					return wc
				}
			}
		}
	}
	return nil
}

// equippedRangedWeapon returns the first equipped weapon with Ranged=true, or nil.
func equippedRangedWeapon(entity *ecs.Entity) *component.WeaponComponent {
	if entity.HasComponent(component.BodyInventory) {
		inv := entity.GetComponent(component.BodyInventory).(*component.BodyInventoryComponent)
		for _, item := range inv.Equipped {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if wc.Ranged {
					return wc
				}
			}
		}
	}
	if entity.HasComponent(component.Inventory) {
		inv := entity.GetComponent(component.Inventory).(*component.InventoryComponent)
		for _, item := range []*ecs.Entity{inv.RightHand, inv.LeftHand} {
			if item != nil && item.HasComponent(component.Weapon) {
				wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
				if wc.Ranged {
					return wc
				}
			}
		}
	}
	return nil
}
