package game

import (
	"fmt"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

const (
	nlModalW = 420
	nlModalH = 300
)

// lootEntry holds one item found near the player.
type lootEntry struct {
	item        *ecs.Entity
	tileX, tileY, tileZ int
	label       string // display string shown in the list
}

// NearbyLootView is the P/E modal that lists items within 1 tile of the player.
// Number keys 1–9 select a row. Pick Up and Equip buttons act on the selection.
type NearbyLootView struct {
	Visible    bool
	OnPickup   func(item *ecs.Entity, tileX, tileY, tileZ int)
	OnEquip    func(item *ecs.Entity, tileX, tileY, tileZ int)

	modal     *minui.Modal
	itemList  *minui.ListBox
	pickupBtn *minui.Button
	equipBtn  *minui.Button
	statusLbl *minui.Label

	entries []lootEntry
}

func NewNearbyLootView() *NearbyLootView {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - nlModalW) / 2
	my := (cfg.ScreenHeight - nlModalH) / 2

	v := &NearbyLootView{}

	v.modal = minui.NewModal("nl_modal", "Nearby Items", nlModalW, nlModalH)
	v.modal.SetPosition(mx, my)
	v.modal.Closeable = true
	v.modal.OnClose = func() { v.Visible = false }

	lbl := minui.NewLabel("nl_lbl", "Select item (1–9 or click):")
	lbl.SetPosition(20, 40)
	v.modal.AddChild(lbl)

	v.itemList = minui.NewListBox("nl_list", []string{})
	v.itemList.SetPosition(20, 62)
	v.itemList.SetSize(nlModalW-40, 140)
	v.itemList.Layout()
	v.modal.AddChild(v.itemList)

	v.statusLbl = minui.NewLabel("nl_status", "")
	v.statusLbl.SetPosition(20, 210)
	v.modal.AddChild(v.statusLbl)

	const btnY = 228
	const btnW = 110
	const btnH = 32

	v.pickupBtn = minui.NewButton("nl_pickup", "Pick Up")
	v.pickupBtn.SetPosition(nlModalW/2-btnW-8, btnY)
	v.pickupBtn.SetSize(btnW, btnH)
	v.pickupBtn.OnClick = func() { v.doPickup() }
	v.modal.AddChild(v.pickupBtn)

	v.equipBtn = minui.NewButton("nl_equip", "Equip")
	v.equipBtn.SetPosition(nlModalW/2+8, btnY)
	v.equipBtn.SetSize(btnW, btnH)
	v.equipBtn.OnClick = func() { v.doEquip() }
	v.modal.AddChild(v.equipBtn)

	return v
}

// Open scans the player's tile and all 8 adjacent tiles for items and shows the modal.
func (v *NearbyLootView) Open(player *ecs.Entity, level *world.Level) {
	v.entries = nil
	v.statusLbl.Text = ""

	pc := player.GetComponent(component.Position).(*component.PositionComponent)
	px, py, pz := pc.GetX(), pc.GetY(), pc.GetZ()

	type offset struct {
		dx, dy int
		label  string
	}
	offsets := []offset{
		{0, 0, "Here"},
		{0, -1, "N"}, {1, -1, "NE"}, {1, 0, "E"}, {1, 1, "SE"},
		{0, 1, "S"}, {-1, 1, "SW"}, {-1, 0, "W"}, {-1, -1, "NW"},
	}

	for _, off := range offsets {
		tx, ty := px+off.dx, py+off.dy
		var buf []*ecs.Entity
		level.Level.GetEntitiesAt(tx, ty, pz, &buf)
		for _, e := range buf {
			if e == player || !e.HasComponent(component.Item) {
				continue
			}
			v.entries = append(v.entries, lootEntry{
				item:  e,
				tileX: tx, tileY: ty, tileZ: pz,
				label: fmt.Sprintf("[%s] %s", off.label, itemName(e)),
			})
		}
	}

	if len(v.entries) == 0 {
		// Nothing nearby — caller should check before opening, but guard here too.
		return
	}

	// Sort so "Here" items always come first, rest by direction label.
	sort.SliceStable(v.entries, func(i, j int) bool {
		iHere := v.entries[i].tileX == px && v.entries[i].tileY == py
		jHere := v.entries[j].tileX == px && v.entries[j].tileY == py
		if iHere != jHere {
			return iHere
		}
		return v.entries[i].label < v.entries[j].label
	})

	labels := make([]string, len(v.entries))
	for i, e := range v.entries {
		hotkey := ""
		if i < 9 {
			hotkey = fmt.Sprintf("%d. ", i+1)
		}
		labels[i] = hotkey + e.label
	}

	v.itemList.SetItems(labels)
	v.modal.SetVisible(true)
	v.Visible = true
}

func (v *NearbyLootView) Update() {
	if !v.Visible {
		return
	}

	numberKeys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
		ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
	}
	for i, k := range numberKeys {
		if inpututil.IsKeyJustPressed(k) && i < len(v.entries) {
			v.itemList.SelectedIndex = i
			break
		}
	}

	v.modal.Update()
}

func (v *NearbyLootView) Draw(screen *ebiten.Image) {
	if !v.Visible {
		return
	}
	v.modal.Draw(screen)
}

func (v *NearbyLootView) doPickup() {
	idx := v.itemList.SelectedIndex
	if idx < 0 || idx >= len(v.entries) {
		v.statusLbl.Text = "Select an item first."
		return
	}
	e := v.entries[idx]
	v.Visible = false
	if v.OnPickup != nil {
		v.OnPickup(e.item, e.tileX, e.tileY, e.tileZ)
	}
}

func (v *NearbyLootView) doEquip() {
	idx := v.itemList.SelectedIndex
	if idx < 0 || idx >= len(v.entries) {
		v.statusLbl.Text = "Select an item first."
		return
	}
	e := v.entries[idx]
	ic := e.item.GetComponent(component.Item).(*component.ItemComponent)
	if ic.Slot == component.BagSlot {
		v.statusLbl.Text = "That item cannot be equipped."
		return
	}
	v.Visible = false
	if v.OnEquip != nil {
		v.OnEquip(e.item, e.tileX, e.tileY, e.tileZ)
	}
}
