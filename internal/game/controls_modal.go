package game

import (
	"image"
	"image/color"
	"log"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/keybindings"
)

const (
	ctrlModalW    = 420
	ctrlModalH    = 580
	ctrlTitleBarH = 30
	ctrlRowH      = 34
	ctrlLabelX    = 20
	ctrlFieldX    = 220
	ctrlFieldW    = 160
	ctrlFieldH    = 26
	ctrlScrollH   = 460 // height of the visible rows area (modal-relative, below title bar)
	ctrlBarW      = 8   // scrollbar width
)

// ctrlRowsTopY is the Y of the first row inside the modal content area (below title bar + gap).
// Children are positioned relative to modal content (title bar already offset by minui).
const ctrlRowsTopY = 8

type controlRow struct {
	action string
	input  *minui.TextInput
}

// ControlsModal lets the player remap all keybindings and save to JSON.
type ControlsModal struct {
	Visible bool

	modal     *minui.Modal
	rows      []controlRow
	scrollOff int // pixel scroll into rows
	statusMsg string
}

func newControlsModal() *ControlsModal {
	cfg := config.Global()
	cx := cfg.ScreenWidth/2 - ctrlModalW/2
	cy := cfg.ScreenHeight/2 - ctrlModalH/2

	m := minui.NewModal("controls_modal", "Controls", ctrlModalW, ctrlModalH)
	m.SetPosition(cx, cy)
	m.Closeable = false

	// Fix modal size — prevent auto-resize from child positions.
	w, h := ctrlModalW, ctrlModalH
	m.GetStyle().Width = &w
	m.GetStyle().Height = &h

	cm := &ControlsModal{modal: m}

	// Build rows from current keybindings.
	kb := keybindings.Global()
	all := kb.All()
	actions := make([]string, 0, len(all))
	for a := range all {
		actions = append(actions, a)
	}
	sort.Strings(actions)

	for i, action := range actions {
		ti := minui.NewTextInput("ctrl_"+action, "")
		ti.Text = all[action]
		ti.SetPosition(ctrlFieldX, ctrlRowsTopY+i*ctrlRowH)
		ti.SetSize(ctrlFieldW, ctrlFieldH)
		m.AddChild(ti)
		cm.rows = append(cm.rows, controlRow{action: action, input: ti})
	}

	// Buttons sit at bottom of modal.
	btnY := ctrlRowsTopY + ctrlScrollH + 10

	saveBtn := minui.NewButton("ctrl_save", "Save")
	saveBtn.SetPosition(ctrlModalW/2-120, btnY)
	saveBtn.SetSize(100, 32)
	saveBtn.OnClick = func() { cm.apply() }
	m.AddChild(saveBtn)

	closeBtn := minui.NewButton("ctrl_close", "Close")
	closeBtn.SetPosition(ctrlModalW/2+20, btnY)
	closeBtn.SetSize(100, 32)
	closeBtn.OnClick = func() { cm.Close() }
	m.AddChild(closeBtn)

	cm.applyScroll()
	return cm
}

func (cm *ControlsModal) Open() {
	cfg := config.Global()
	cm.modal.SetPosition(cfg.ScreenWidth/2-ctrlModalW/2, cfg.ScreenHeight/2-ctrlModalH/2)

	// Refresh from current bindings (may have been saved since modal was built).
	kb := keybindings.Global()
	all := kb.All()
	for _, row := range cm.rows {
		if v, ok := all[row.action]; ok {
			row.input.Text = v
		}
	}

	cm.scrollOff = 0
	cm.statusMsg = ""
	cm.applyScroll()
	cm.modal.SetVisible(true)
	cm.Visible = true
}

func (cm *ControlsModal) Close() {
	cm.modal.SetVisible(false)
	cm.Visible = false
}

// applyScroll repositions inputs and shows/hides them based on current scroll offset.
func (cm *ControlsModal) applyScroll() {
	for i, row := range cm.rows {
		y := ctrlRowsTopY + i*ctrlRowH - cm.scrollOff
		row.input.SetPosition(ctrlFieldX, y)
		inView := y+ctrlFieldH > ctrlRowsTopY && y < ctrlRowsTopY+ctrlScrollH
		row.input.SetVisible(inView)
		row.input.SetEnabled(inView)
	}
}

func (cm *ControlsModal) maxScroll() int {
	max := len(cm.rows)*ctrlRowH - ctrlScrollH
	if max < 0 {
		return 0
	}
	return max
}

func (cm *ControlsModal) scroll(delta int) {
	cm.scrollOff += delta
	if cm.scrollOff < 0 {
		cm.scrollOff = 0
	}
	if cm.scrollOff > cm.maxScroll() {
		cm.scrollOff = cm.maxScroll()
	}
	cm.applyScroll()
}

func (cm *ControlsModal) apply() {
	kb := keybindings.Global()
	for _, row := range cm.rows {
		if row.input.Text != "" {
			kb.Set(row.action, row.input.Text)
		}
	}
	if err := kb.Save(); err != nil {
		log.Printf("controls save failed: %v", err)
		cm.statusMsg = "Save failed."
	} else {
		cm.statusMsg = "Saved."
	}
}

func (cm *ControlsModal) Update() {
	if !cm.Visible {
		return
	}

	_, dy := ebiten.Wheel()
	if dy != 0 {
		cm.scroll(-int(dy) * ctrlRowH)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		cm.scroll(ctrlRowH)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		cm.scroll(-ctrlRowH)
	}

	cm.modal.Update()
}

func (cm *ControlsModal) Draw(screen *ebiten.Image) {
	if !cm.Visible {
		return
	}
	cm.modal.Draw(screen)

	cfg := config.Global()
	cx := cfg.ScreenWidth/2 - ctrlModalW/2
	cy := cfg.ScreenHeight/2 - ctrlModalH/2

	// absRowsTop accounts for the title bar that minui adds above child positions.
	absRowsTop := cy + ctrlTitleBarH + ctrlRowsTopY
	absRowsBot := absRowsTop + ctrlScrollH

	// Clip to scroll region so labels don't bleed outside it.
	clipRect := image.Rect(cx, absRowsTop, cx+ctrlModalW-ctrlBarW-4, absRowsBot)
	clip := screen.SubImage(clipRect).(*ebiten.Image)

	labelColor := color.RGBA{180, 200, 180, 255}
	const fontSize = 12

	for i, row := range cm.rows {
		// Absolute Y of vertical center of this row.
		absY := absRowsTop + i*ctrlRowH - cm.scrollOff + (ctrlRowH-fontSize)/2
		if absY+fontSize <= absRowsTop || absY >= absRowsBot {
			continue
		}
		mlge_text.Draw(clip, row.action, fontSize, cx+ctrlLabelX, absY, labelColor)
	}

	// Scrollbar track.
	trackX := cx + ctrlModalW - ctrlBarW - 4
	trackColor := color.RGBA{50, 55, 60, 255}
	drawFilledRect(screen, trackX, absRowsTop, ctrlBarW, ctrlScrollH, trackColor)

	// Scrollbar thumb.
	total := len(cm.rows) * ctrlRowH
	if total > ctrlScrollH {
		thumbH := ctrlScrollH * ctrlScrollH / total
		if thumbH < 16 {
			thumbH = 16
		}
		thumbY := absRowsTop + cm.scrollOff*(ctrlScrollH-thumbH)/cm.maxScroll()
		thumbColor := color.RGBA{120, 140, 180, 220}
		drawFilledRect(screen, trackX, thumbY, ctrlBarW, thumbH, thumbColor)
	}

	if cm.statusMsg != "" {
		mlge_text.Draw(screen, cm.statusMsg, 12, cx+20, absRowsBot+8, color.RGBA{120, 200, 120, 255})
	}
}

func drawFilledRect(screen *ebiten.Image, x, y, w, h int, c color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	rect := ebiten.NewImage(w, h)
	rect.Fill(c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(rect, op)
}
