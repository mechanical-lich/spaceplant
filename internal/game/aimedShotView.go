package game

import (
	"fmt"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

const (
	asModalW = 320
	asModalH = 240
)

// AimedShotView is the modal shown when the player presses Shift+F.
// It lists the target entity's non-amputated body parts and lets the player
// choose one for a targeted aimed shot. Number keys 1–9 act as hotkeys.
type AimedShotView struct {
	Visible  bool
	OnSelect func(bodyPart string)

	modal    *minui.Modal
	partList *minui.ListBox

	parts []string // parallel to partList items
}

func NewAimedShotView() *AimedShotView {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - asModalW) / 2
	my := (cfg.ScreenHeight - asModalH) / 2

	v := &AimedShotView{}

	v.modal = minui.NewModal("as_modal", "Aimed Shot — Choose Target", asModalW, asModalH)
	v.modal.SetPosition(mx, my)
	v.modal.Closeable = true
	v.modal.OnClose = func() { v.Visible = false }

	lbl := minui.NewLabel("as_lbl", "Select body part (1–9 or click):")
	lbl.SetPosition(20, 40)
	v.modal.AddChild(lbl)

	v.partList = minui.NewListBox("as_list", []string{})
	v.partList.SetPosition(20, 65)
	v.partList.SetSize(asModalW-40, 140)
	v.partList.Layout()
	v.partList.OnSelect = func(idx int, _ string) {
		v.confirm(idx)
	}
	v.modal.AddChild(v.partList)

	return v
}

// Open populates the list from the target entity and shows the modal.
func (v *AimedShotView) Open(target *ecs.Entity) {
	v.parts = nil
	var labels []string

	if target != nil && target.HasComponent(rlcomponents.Body) {
		bc := target.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
		var names []string
		for name, part := range bc.Parts {
			if !part.Amputated {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for i, name := range names {
			hotkey := ""
			if i < 9 {
				hotkey = fmt.Sprintf("[%d] ", i+1)
			}
			labels = append(labels, hotkey+name)
			v.parts = append(v.parts, name)
		}
	}

	v.partList.SetItems(labels)
	v.modal.SetVisible(true)
	v.Visible = true
}

func (v *AimedShotView) Update() {
	if !v.Visible {
		return
	}

	// Number key hotkeys 1–9.
	numberKeys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
		ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
	}
	for i, k := range numberKeys {
		if inpututil.IsKeyJustPressed(k) && i < len(v.parts) {
			v.confirm(i)
			return
		}
	}

	v.modal.Update()
}

func (v *AimedShotView) Draw(screen *ebiten.Image) {
	if !v.Visible {
		return
	}
	v.modal.Draw(screen)
}

func (v *AimedShotView) confirm(idx int) {
	if idx < 0 || idx >= len(v.parts) {
		return
	}
	part := v.parts[idx]
	v.Visible = false
	if v.OnSelect != nil {
		v.OnSelect(part)
	}
}
