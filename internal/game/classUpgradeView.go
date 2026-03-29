package game

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/ecs"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

const (
	classModalW = 460
	classModalH = 360
)

// ClassUpgradeView is the Shift+C class upgrade modal.
type ClassUpgradeView struct {
	Visible bool
	player  *ecs.Entity

	modal       *minui.Modal
	list        *minui.ListBox
	descArea    *minui.ScrollingTextArea
	pointsLabel *minui.Label
	buyButton   *minui.Button

	skillIDs    []string // parallel index → skill ID
	selectedIdx int
}

func NewClassUpgradeView(player *ecs.Entity) *ClassUpgradeView {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - classModalW) / 2
	my := (cfg.ScreenHeight - classModalH) / 2

	v := &ClassUpgradeView{player: player, selectedIdx: -1}

	v.modal = minui.NewModal("class_modal", "Class Upgrades", classModalW, classModalH)
	v.modal.SetPosition(mx, my)
	v.modal.Closeable = true
	v.modal.OnClose = func() { v.Visible = false }

	// Left: skill list (positions relative to modal content area)
	v.list = minui.NewListBox("class_skills", []string{})
	v.list.SetPosition(10, 10)
	v.list.SetSize(180, 255)
	v.list.Layout()
	v.list.OnSelect = func(idx int, _ string) {
		v.selectedIdx = idx
		v.refreshDesc()
	}
	v.modal.AddChild(v.list)

	// Right: scrolling description area
	v.descArea = minui.NewScrollingTextArea("class_desc", 250, 210)
	v.descArea.SetPosition(200, 10)
	v.descArea.LineHeight = 14
	v.modal.AddChild(v.descArea)

	// Bottom-left: points remaining
	v.pointsLabel = minui.NewLabel("class_points", "")
	v.pointsLabel.SetPosition(10, 275)
	v.pointsLabel.SetSize(200, 20)
	v.modal.AddChild(v.pointsLabel)

	// Bottom-right: buy button
	v.buyButton = minui.NewButton("class_buy", "Buy Skill")
	v.buyButton.SetPosition(320, 270)
	v.buyButton.SetSize(120, 30)
	v.buyButton.OnClick = func() {
		if v.selectedIdx < 0 || v.selectedIdx >= len(v.skillIDs) {
			return
		}
		if class.BuySkill(v.player, v.skillIDs[v.selectedIdx]) {
			v.refreshAll()
		}
	}
	v.modal.AddChild(v.buyButton)

	return v
}

// Open refreshes data and makes the view visible.
func (v *ClassUpgradeView) Open() {
	v.refreshAll()
	v.modal.SetVisible(true)
	v.Visible = true
}

func (v *ClassUpgradeView) Update() {
	if !v.Visible {
		return
	}
	v.modal.Update()
}

func (v *ClassUpgradeView) Draw(screen *ebiten.Image) {
	if !v.Visible {
		return
	}
	v.modal.Draw(screen)
}

// refreshAll rebuilds the skill list and points display.
func (v *ClassUpgradeView) refreshAll() {
	if !v.player.HasComponent(component.Class) {
		return
	}
	cc := v.player.GetComponent(component.Class).(*component.ClassComponent)

	v.skillIDs = v.skillIDs[:0]
	items := []string{}
	seen := map[string]bool{}

	for _, classID := range cc.Classes {
		def := class.Get(classID)
		if def == nil {
			continue
		}
		for _, sID := range def.Skills {
			if seen[sID] {
				continue
			}
			seen[sID] = true
			v.skillIDs = append(v.skillIDs, sID)

			name := sID
			if sd := skill.Get(sID); sd != nil {
				name = sd.Name
			}
			if slices.Contains(cc.ChosenSkills, sID) {
				name = "✓ " + name
			}
			items = append(items, name)
		}
	}

	v.list.SetItems(items)
	if v.selectedIdx >= len(items) {
		v.selectedIdx = -1
		v.list.SelectedIndex = -1
	}

	v.pointsLabel.Text = fmt.Sprintf("Upgrade Points: %d", cc.UpgradePoints)
	v.refreshDesc()
}

// refreshDesc updates the description area for the currently selected skill.
func (v *ClassUpgradeView) refreshDesc() {
	// clear by replacing — ScrollingTextArea doesn't have a clear, so rebuild
	v.descArea.Clear()

	if v.selectedIdx < 0 || v.selectedIdx >= len(v.skillIDs) {
		v.descArea.AddText("Select a skill to see its description.")
		return
	}

	sd := skill.Get(v.skillIDs[v.selectedIdx])
	if sd == nil {
		return
	}

	v.descArea.AddText(sd.Name)
	v.descArea.AddText("")
	for _, line := range mlge_text.Wrap(sd.Description, 30, 0) {
		v.descArea.AddText(line)
	}

	if v.player.HasComponent(component.Class) {
		cc := v.player.GetComponent(component.Class).(*component.ClassComponent)
		if slices.Contains(cc.ChosenSkills, v.skillIDs[v.selectedIdx]) {
			v.descArea.AddText("")
			v.descArea.AddText("(already purchased)")
		}
	}

}
