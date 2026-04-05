package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

const (
	rlModalW = 500
	rlModalH = 360
)

// ReloadView is the Shift+R reload modal.
// Left list shows equipped ranged weapons; right list shows compatible ammo in bag.
// Clicking "Reload" loads the selected ammo into the selected weapon.
type ReloadView struct {
	Visible  bool
	OnReload func(weaponItem, ammoItem *ecs.Entity)
	player   *ecs.Entity

	modal      *minui.Modal
	weaponList *minui.ListBox
	ammoList   *minui.ListBox
	reloadBtn  *minui.Button
	statusLbl  *minui.Label

	weapons []*ecs.Entity // parallel to weaponList items
	ammos   []*ecs.Entity // parallel to ammoList items (filtered by selected weapon)
}

func NewReloadView(player *ecs.Entity) *ReloadView {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - rlModalW) / 2
	my := (cfg.ScreenHeight - rlModalH) / 2

	v := &ReloadView{player: player}

	v.modal = minui.NewModal("rl_modal", "Reload Weapon", rlModalW, rlModalH)
	v.modal.SetPosition(mx, my)
	v.modal.Closeable = true
	v.modal.OnClose = func() { v.Visible = false }

	const listW = 190
	const listH = 220
	const gap = 20
	const startY = 40

	// Weapon list (left)
	weaponLbl := minui.NewLabel("rl_wlbl", "Weapon")
	weaponLbl.SetPosition(20, startY)
	v.modal.AddChild(weaponLbl)

	v.weaponList = minui.NewListBox("rl_wlist", []string{})
	v.weaponList.SetPosition(20, startY+20)
	v.weaponList.SetSize(listW, listH)
	v.weaponList.Layout()
	v.weaponList.OnSelect = func(_ int, _ string) { v.refreshAmmoList() }
	v.modal.AddChild(v.weaponList)

	// Ammo list (right)
	ammoLbl := minui.NewLabel("rl_albl", "Compatible Ammo (in bag)")
	ammoLbl.SetPosition(20+listW+gap, startY)
	v.modal.AddChild(ammoLbl)

	v.ammoList = minui.NewListBox("rl_alist", []string{})
	v.ammoList.SetPosition(20+listW+gap, startY+20)
	v.ammoList.SetSize(listW, listH)
	v.ammoList.Layout()
	v.modal.AddChild(v.ammoList)

	// Status label
	v.statusLbl = minui.NewLabel("rl_status", "")
	v.statusLbl.SetPosition(20, startY+listH+30)
	v.modal.AddChild(v.statusLbl)

	// Reload button
	v.reloadBtn = minui.NewButton("rl_btn", "Reload")
	v.reloadBtn.SetPosition(rlModalW/2-50, startY+listH+55)
	v.reloadBtn.SetSize(100, 32)
	v.reloadBtn.OnClick = func() { v.doReload() }
	v.modal.AddChild(v.reloadBtn)

	return v
}

// Open refreshes data and shows the modal.
func (v *ReloadView) Open() {
	v.refreshWeaponList()
	v.refreshAmmoList()
	v.statusLbl.Text = ""
	v.modal.SetVisible(true)
	v.Visible = true
}

func (v *ReloadView) Update() {
	if !v.Visible {
		return
	}
	v.modal.Update()
}

func (v *ReloadView) Draw(screen *ebiten.Image) {
	if !v.Visible {
		return
	}
	v.modal.Draw(screen)
}

// refreshWeaponList populates the weapon list with all equipped ranged weapons.
func (v *ReloadView) refreshWeaponList() {
	v.weapons = nil
	var labels []string

	equipped := playerEquipped(v.player)
	for _, item := range equipped {
		if item == nil || !item.HasComponent(component.Weapon) {
			continue
		}
		wc := item.GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
		if !wc.Ranged || wc.MaxMagazine == 0 {
			continue
		}
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		label := fmt.Sprintf("%s [%d/%d]", ic.Name, wc.Magazine, wc.MaxMagazine)
		labels = append(labels, label)
		v.weapons = append(v.weapons, item)
	}

	v.weaponList.SetItems(labels)
}

// refreshAmmoList re-populates the ammo list based on the selected weapon's AmmoType.
func (v *ReloadView) refreshAmmoList() {
	v.ammos = nil
	var labels []string

	idx := v.weaponList.SelectedIndex
	var ammoType string
	if idx >= 0 && idx < len(v.weapons) {
		wc := v.weapons[idx].GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
		ammoType = wc.AmmoType
	}

	bag := playerBag(v.player)
	for _, item := range bag {
		if item == nil || !item.HasComponent(component.Ammo) {
			continue
		}
		ac := item.GetComponent(component.Ammo).(*component.AmmoComponent)
		if ammoType != "" && ac.AmmoType != ammoType {
			continue
		}
		ic := item.GetComponent(component.Item).(*component.ItemComponent)
		label := fmt.Sprintf("%s (×%d)", ic.Name, ac.Count)
		labels = append(labels, label)
		v.ammos = append(v.ammos, item)
	}

	v.ammoList.SetItems(labels)
}

// doReload validates the selection and delegates the actual ECS mutation to OnReload,
// which routes the reload through the sim under its write lock.
func (v *ReloadView) doReload() {
	wIdx := v.weaponList.SelectedIndex
	aIdx := v.ammoList.SelectedIndex

	if wIdx < 0 || wIdx >= len(v.weapons) {
		v.statusLbl.Text = "Select a weapon first."
		return
	}
	if aIdx < 0 || aIdx >= len(v.ammos) {
		v.statusLbl.Text = "Select ammo first."
		return
	}

	weaponItem := v.weapons[wIdx]
	ammoItem := v.ammos[aIdx]

	wc := weaponItem.GetComponent(component.Weapon).(*rlcomponents.WeaponComponent)
	ac := ammoItem.GetComponent(component.Ammo).(*component.AmmoComponent)

	if ac.AmmoType != wc.AmmoType {
		v.statusLbl.Text = "Incompatible ammo type."
		return
	}
	if wc.MaxMagazine-wc.Magazine <= 0 {
		v.statusLbl.Text = "Magazine already full."
		return
	}

	if v.OnReload != nil {
		v.OnReload(weaponItem, ammoItem)
	}

	v.Visible = false
}
