package game

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// ─── shared helpers ────────────────────────────────────────────────────────

func termDrawModal(s tcell.Screen, title string, lines []string, hint string) {
	sw, sh := s.Size()
	w := sw * 2 / 3
	if w < 50 {
		w = 50
	}
	h := len(lines) + 6
	if h < 10 {
		h = 10
	}
	if h > sh-4 {
		h = sh - 4
	}
	x := (sw - w) / 2
	y := (sh - h) / 2

	border := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	content := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	hintSt := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)

	rltermgui.FillRect(s, x, y, w, h, content)
	rltermgui.DrawBox(s, x, y, w, h, " "+title+" ", border)

	for i, line := range lines {
		if y+1+i >= y+h-2 {
			break
		}
		if len([]rune(line)) > w-2 {
			line = string([]rune(line)[:w-2])
		}
		rltermgui.DrawText(s, x+1, y+1+i, line, content)
	}
	if hint != "" {
		rltermgui.DrawText(s, x+1, y+h-2, hint, hintSt)
	}
}

// ─── TermReloadView ─────────────────────────────────────────────────────────

// TermReloadView is the terminal reload modal.
// Left column: ranged weapons with magazines. Right column: matching ammo.
type TermReloadView struct {
	visible  bool
	OnReload func(weaponItem, ammoItem *ecs.Entity)

	player      *ecs.Entity
	weapons     []*ecs.Entity
	ammos       []*ecs.Entity
	weaponCursor int
	ammoCursor   int
	focusAmmo   bool // false = focus on weapons, true = focus on ammo
	status      string
}

func NewTermReloadView(player *ecs.Entity) *TermReloadView {
	return &TermReloadView{player: player}
}

func (v *TermReloadView) Visible() bool { return v.visible }

func (v *TermReloadView) Open() {
	v.weapons = nil
	v.ammos = nil
	v.weaponCursor = 0
	v.ammoCursor = 0
	v.focusAmmo = false
	v.status = ""

	bag := playerBag(v.player)
	for _, item := range bag {
		if !item.HasComponent(component.Weapon) {
			continue
		}
		wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
		if wc.Ranged && wc.MaxMagazine > 0 {
			v.weapons = append(v.weapons, item)
		}
	}
	v.refreshAmmo()
	v.visible = true
}

func (v *TermReloadView) refreshAmmo() {
	v.ammos = nil
	v.ammoCursor = 0
	ammoType := ""
	if v.weaponCursor < len(v.weapons) {
		wc := v.weapons[v.weaponCursor].GetComponent(component.Weapon).(*component.WeaponComponent)
		ammoType = wc.AmmoType
	}
	bag := playerBag(v.player)
	for _, item := range bag {
		if !item.HasComponent(component.Ammo) {
			continue
		}
		if ammoType != "" {
			ac := item.GetComponent(component.Ammo).(*component.AmmoComponent)
			if ac.AmmoType != ammoType {
				continue
			}
		}
		v.ammos = append(v.ammos, item)
	}
}

func (v *TermReloadView) HandleKey(ev *tcell.EventKey) bool {
	if !v.visible {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		v.visible = false
	case tcell.KeyTab:
		v.focusAmmo = !v.focusAmmo
	case tcell.KeyUp:
		if !v.focusAmmo {
			if v.weaponCursor > 0 {
				v.weaponCursor--
				v.refreshAmmo()
			}
		} else {
			if v.ammoCursor > 0 {
				v.ammoCursor--
			}
		}
	case tcell.KeyDown:
		if !v.focusAmmo {
			if v.weaponCursor < len(v.weapons)-1 {
				v.weaponCursor++
				v.refreshAmmo()
			}
		} else {
			if v.ammoCursor < len(v.ammos)-1 {
				v.ammoCursor++
			}
		}
	case tcell.KeyEnter:
		v.doReload()
	}
	return true
}

func (v *TermReloadView) doReload() {
	if v.weaponCursor >= len(v.weapons) {
		v.status = "No weapon selected."
		return
	}
	if v.ammoCursor >= len(v.ammos) {
		v.status = "No ammo selected."
		return
	}
	weaponItem := v.weapons[v.weaponCursor]
	ammoItem := v.ammos[v.ammoCursor]
	wc := weaponItem.GetComponent(component.Weapon).(*component.WeaponComponent)
	ac := ammoItem.GetComponent(component.Ammo).(*component.AmmoComponent)
	if ac.AmmoType != wc.AmmoType {
		v.status = "Wrong ammo type."
		return
	}
	if wc.MaxMagazine-wc.Magazine <= 0 {
		v.status = "Magazine already full."
		return
	}
	if v.OnReload != nil {
		v.OnReload(weaponItem, ammoItem)
	}
	v.visible = false
}

func (v *TermReloadView) Draw(s tcell.Screen) {
	if !v.visible {
		return
	}
	sw, sh := s.Size()
	w := sw * 2 / 3
	if w < 50 {
		w = 50
	}
	h := 20
	x := (sw - w) / 2
	y := (sh - h) / 2

	border := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	bg := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	sel := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	hint := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)

	rltermgui.FillRect(s, x, y, w, h, bg)
	rltermgui.DrawBox(s, x, y, w, h, " Reload ", border)

	colW := (w - 2) / 2
	// Weapon column header
	wHeader := "Weapons"
	if !v.focusAmmo {
		wHeader = "[Weapons]"
	}
	rltermgui.DrawText(s, x+1, y+1, wHeader, dim)
	// Ammo column header
	aHeader := "Ammo"
	if v.focusAmmo {
		aHeader = "[Ammo]"
	}
	rltermgui.DrawText(s, x+1+colW, y+1, aHeader, dim)

	maxRows := h - 5
	for i, item := range v.weapons {
		if i >= maxRows {
			break
		}
		wc := item.GetComponent(component.Weapon).(*component.WeaponComponent)
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		label := fmt.Sprintf("%s [%d/%d]", ic.Name, wc.Magazine, wc.MaxMagazine)
		if len([]rune(label)) > colW-1 {
			label = string([]rune(label)[:colW-1])
		}
		st := bg
		if i == v.weaponCursor && !v.focusAmmo {
			st = sel
		} else if i == v.weaponCursor {
			st = dim
		}
		rltermgui.DrawText(s, x+1, y+2+i, label, st)
	}
	if len(v.weapons) == 0 {
		rltermgui.DrawText(s, x+1, y+2, "(none)", dim)
	}

	for i, item := range v.ammos {
		if i >= maxRows {
			break
		}
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		label := ic.Name
		if len([]rune(label)) > colW-1 {
			label = string([]rune(label)[:colW-1])
		}
		st := bg
		if i == v.ammoCursor && v.focusAmmo {
			st = sel
		} else if i == v.ammoCursor {
			st = dim
		}
		rltermgui.DrawText(s, x+1+colW, y+2+i, label, st)
	}
	if len(v.ammos) == 0 {
		rltermgui.DrawText(s, x+1+colW, y+2, "(none)", dim)
	}

	if v.status != "" {
		rltermgui.DrawText(s, x+1, y+h-3, v.status, dim)
	}
	rltermgui.DrawText(s, x+1, y+h-2, "[Tab] switch column  [↑↓] select  [Enter] reload  [Esc] close", hint)
}

// ─── TermAimedShotView ──────────────────────────────────────────────────────

type TermAimedShotView struct {
	visible  bool
	OnSelect func(bodyPart string)

	parts  []string
	cursor int
}

func NewTermAimedShotView() *TermAimedShotView  { return &TermAimedShotView{} }
func (v *TermAimedShotView) Visible() bool      { return v.visible }

func (v *TermAimedShotView) Open(target *ecs.Entity) {
	v.parts = nil
	v.cursor = 0
	if target != nil && target.HasComponent(rlcomponents.Body) {
		bc := target.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
		var names []string
		for name, part := range bc.Parts {
			if !part.Amputated {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		v.parts = names
	}
	v.visible = true
}

func (v *TermAimedShotView) HandleKey(ev *tcell.EventKey) bool {
	if !v.visible {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		v.visible = false
	case tcell.KeyUp:
		if v.cursor > 0 {
			v.cursor--
		}
	case tcell.KeyDown:
		if v.cursor < len(v.parts)-1 {
			v.cursor++
		}
	case tcell.KeyEnter:
		if v.cursor < len(v.parts) {
			if v.OnSelect != nil {
				v.OnSelect(v.parts[v.cursor])
			}
			v.visible = false
		}
	case tcell.KeyRune:
		r := ev.Rune()
		if r >= '1' && r <= '9' {
			idx := int(r-'1')
			if idx < len(v.parts) {
				if v.OnSelect != nil {
					v.OnSelect(v.parts[idx])
				}
				v.visible = false
			}
		}
	}
	return true
}

func (v *TermAimedShotView) Draw(s tcell.Screen) {
	if !v.visible {
		return
	}
	var lines []string
	for i, part := range v.parts {
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		hotkey := ""
		if i < 9 {
			hotkey = fmt.Sprintf("[%d] ", i+1)
		}
		lines = append(lines, prefix+hotkey+part)
	}
	if len(lines) == 0 {
		lines = []string{"  (no valid targets)"}
	}
	termDrawModal(s, "Aimed Shot — Choose Target", lines, "[↑↓] select  [1-9] hotkey  [Enter] confirm  [Esc] cancel")
}

// ─── TermLootView ───────────────────────────────────────────────────────────

type termLootEntry struct {
	item              *ecs.Entity
	tileX, tileY, tileZ int
}

type TermLootView struct {
	visible  bool
	OnPickup func(item *ecs.Entity, tileX, tileY, tileZ int)
	OnEquip  func(item *ecs.Entity, tileX, tileY, tileZ int)

	entries []termLootEntry
	cursor  int
	status  string
}

func NewTermLootView() *TermLootView  { return &TermLootView{} }
func (v *TermLootView) Visible() bool { return v.visible }

func (v *TermLootView) Open(player *ecs.Entity, level *world.Level) {
	v.entries = nil
	v.cursor = 0
	v.status = ""
	if player == nil {
		return
	}
	pc := player.GetComponent("Position").(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()
	seen := map[*ecs.Entity]bool{}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			var buf []*ecs.Entity
			level.Level.GetEntitiesAt(px+dx, py+dy, pz, &buf)
			for _, e := range buf {
				if e == player || seen[e] || !e.HasComponent(component.Item) {
					continue
				}
				seen[e] = true
				v.entries = append(v.entries, termLootEntry{
					item: e, tileX: px + dx, tileY: py + dy, tileZ: pz,
				})
			}
		}
	}
	v.visible = true
}

func (v *TermLootView) HandleKey(ev *tcell.EventKey) bool {
	if !v.visible {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		v.visible = false
	case tcell.KeyUp:
		if v.cursor > 0 {
			v.cursor--
		}
	case tcell.KeyDown:
		if v.cursor < len(v.entries)-1 {
			v.cursor++
		}
	case tcell.KeyEnter, tcell.KeyRune:
		r := ev.Rune()
		if ev.Key() == tcell.KeyEnter || r == 'p' || r == 'P' {
			v.doPickup()
		} else if r == 'e' || r == 'E' {
			v.doEquip()
		} else if r >= '1' && r <= '9' {
			idx := int(r - '1')
			if idx < len(v.entries) {
				v.cursor = idx
			}
		}
	}
	return true
}

func (v *TermLootView) doPickup() {
	if v.cursor >= len(v.entries) {
		return
	}
	e := v.entries[v.cursor]
	if v.OnPickup != nil {
		v.OnPickup(e.item, e.tileX, e.tileY, e.tileZ)
	}
	v.visible = false
}

func (v *TermLootView) doEquip() {
	if v.cursor >= len(v.entries) {
		return
	}
	e := v.entries[v.cursor]
	if v.OnEquip != nil {
		v.OnEquip(e.item, e.tileX, e.tileY, e.tileZ)
	}
	v.visible = false
}

func (v *TermLootView) Draw(s tcell.Screen) {
	if !v.visible {
		return
	}
	var lines []string
	for i, e := range v.entries {
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		hotkey := ""
		if i < 9 {
			hotkey = fmt.Sprintf("[%d] ", i+1)
		}
		lines = append(lines, prefix+hotkey+itemName(e.item))
	}
	if len(lines) == 0 {
		lines = []string{"  (nothing nearby)"}
	}
	if v.status != "" {
		lines = append(lines, "", v.status)
	}
	termDrawModal(s, "Nearby Items", lines, "[↑↓/1-9] select  [Enter/P] pick up  [E] equip  [Esc] close")
}

// ─── TermClassView ──────────────────────────────────────────────────────────

type TermClassView struct {
	visible bool
	player  *ecs.Entity

	skillIDs []string
	cursor   int
	status   string
}

func NewTermClassView(player *ecs.Entity) *TermClassView {
	return &TermClassView{player: player}
}

func (v *TermClassView) Visible() bool { return v.visible }

func (v *TermClassView) Open() {
	v.cursor = 0
	v.status = ""
	v.skillIDs = nil

	if !v.player.HasComponent(component.Class) {
		v.status = "No class assigned."
		v.visible = true
		return
	}
	cc := v.player.GetComponent(component.Class).(*component.ClassComponent)
	existing := map[string]bool{}
	if v.player.HasComponent(component.Skill) {
		sc := v.player.GetComponent(component.Skill).(*component.SkillComponent)
		for _, s := range sc.Skills {
			existing[s] = true
		}
	}
	seen := map[string]bool{}
	for _, classID := range cc.Classes {
		cl := class.Get(classID)
		if cl == nil {
			continue
		}
		for _, sid := range cl.Skills {
			if !existing[sid] && !seen[sid] {
				seen[sid] = true
				v.skillIDs = append(v.skillIDs, sid)
			}
		}
	}
	slices.Sort(v.skillIDs)
	v.visible = true
}

func (v *TermClassView) getUpgradePoints() int {
	if !v.player.HasComponent(component.Class) {
		return 0
	}
	cc := v.player.GetComponent(component.Class).(*component.ClassComponent)
	return cc.UpgradePoints
}

func (v *TermClassView) HandleKey(ev *tcell.EventKey) bool {
	if !v.visible {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		v.visible = false
	case tcell.KeyUp:
		if v.cursor > 0 {
			v.cursor--
		}
	case tcell.KeyDown:
		if v.cursor < len(v.skillIDs)-1 {
			v.cursor++
		}
	case tcell.KeyEnter:
		v.buySkill()
	}
	return true
}

func (v *TermClassView) buySkill() {
	if v.cursor >= len(v.skillIDs) {
		v.status = "Nothing selected."
		return
	}
	pts := v.getUpgradePoints()
	if pts <= 0 {
		v.status = "No upgrade points remaining."
		return
	}
	sid := v.skillIDs[v.cursor]
	cc := v.player.GetComponent(component.Class).(*component.ClassComponent)
	cc.UpgradePoints--
	skill.Apply(v.player, sid)
	// Remove from available list
	v.skillIDs = append(v.skillIDs[:v.cursor], v.skillIDs[v.cursor+1:]...)
	if v.cursor >= len(v.skillIDs) && v.cursor > 0 {
		v.cursor--
	}
	v.status = fmt.Sprintf("Purchased %s.", sid)
}

func (v *TermClassView) Draw(s tcell.Screen) {
	if !v.visible {
		return
	}
	pts := v.getUpgradePoints()
	var lines []string
	lines = append(lines, fmt.Sprintf("Upgrade points: %d", pts), "")
	for i, sid := range v.skillIDs {
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		def := skill.Get(sid)
		name := sid
		desc := ""
		if def != nil {
			name = def.Name
			desc = def.Description
		}
		line := prefix + name
		if desc != "" {
			line += " — " + desc
		}
		lines = append(lines, line)
	}
	if len(v.skillIDs) == 0 {
		lines = append(lines, "  (no skills available)")
	}
	if v.status != "" {
		lines = append(lines, "", v.status)
	}
	hint := "[↑↓] select  [Enter] buy  [Esc] close"
	termDrawModal(s, "Class Upgrades", lines, hint)
}

// ─── TermRayTarget ──────────────────────────────────────────────────────────

// TermRayTarget walks the player's facing direction and returns the first solid
// non-door (or open-door-skipped) entity within range, or nil.
func TermRayTarget(player *ecs.Entity, level *world.Level) *ecs.Entity {
	if player == nil {
		return nil
	}
	pc := player.GetComponent("Position").(*component.PositionComponent)
	x, y, z := pc.GetX(), pc.GetY(), pc.GetZ()
	dx, dy := 0, -1
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
		if level.IsTileSolid(tx, ty, z) {
			break
		}
		e := level.Level.GetSolidEntityAt(tx, ty, z)
		if e != nil && e != player {
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

// TermHasNearbyItems reports whether any item entities exist within 1 tile of the player.
func TermHasNearbyItems(player *ecs.Entity, level *world.Level) bool {
	if player == nil {
		return false
	}
	pc := player.GetComponent("Position").(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			var buf []*ecs.Entity
			level.Level.GetEntitiesAt(px+dx, py+dy, pz, &buf)
			for _, e := range buf {
				if e != player && e.HasComponent(component.Item) {
					return true
				}
			}
		}
	}
	return false
}

// TermSkillNames returns skill display names for a player, comma-separated, for the HUD.
func TermSkillNames(player *ecs.Entity) string {
	if !player.HasComponent(component.Skill) {
		return ""
	}
	sc := player.GetComponent(component.Skill).(*component.SkillComponent)
	var names []string
	for _, sid := range sc.Skills {
		def := skill.Get(sid)
		if def != nil {
			names = append(names, def.Name)
		} else {
			names = append(names, sid)
		}
	}
	slices.Sort(names)
	return strings.Join(names, ", ")
}
