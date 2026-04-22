package game

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// OptionsModal lets the player edit a subset of config values and save them.
type OptionsModal struct {
	modal *minui.Modal

	crtInput         *minui.TextInput
	pressDelayInput  *minui.TextInput
	renderScaleInput *minui.TextInput
	npcDelayInput    *minui.TextInput
	fullscreenToggle *minui.Toggle

	saveBtn   *minui.Button
	statusMsg string

	Visible bool
}

func newOptionsModal() *OptionsModal {
	cfg := config.Global()

	const (
		modalW = 360
		modalH = 368
		fieldX = 180
		fieldW = 140
		fieldH = 28
		startY = 50
		rowGap = 44
	)

	cx := cfg.ScreenWidth/2 - modalW/2
	cy := cfg.ScreenHeight/2 - modalH/2

	m := minui.NewModal("options_modal", "Options", modalW, modalH)
	m.SetPosition(cx, cy)
	m.Closeable = false

	makeInput := func(id, val string, row int) *minui.TextInput {
		ti := minui.NewTextInput(id, "")
		ti.Text = val
		ti.SetPosition(fieldX, startY+row*rowGap)
		ti.SetSize(fieldW, fieldH)
		m.AddChild(ti)
		return ti
	}

	crtInput := makeInput("opt_crt", fmt.Sprintf("%.2f", cfg.CRTIntensity), 0)
	pressDelayInput := makeInput("opt_press_delay", strconv.Itoa(cfg.PressDelay), 1)
	renderScaleInput := makeInput("opt_render_scale", fmt.Sprintf("%.1f", cfg.RenderScale), 2)
	npcDelayInput := makeInput("opt_npc_delay", strconv.Itoa(cfg.NpcTurnDelayTicks), 3)

	fsToggle := minui.NewToggle("opt_fullscreen", "")
	fsToggle.On = cfg.Fullscreen
	fsToggle.SetPosition(fieldX, startY+4*rowGap)
	fsToggle.SetSize(fieldH*2, fieldH) // compact square-ish toggle
	m.AddChild(fsToggle)

	saveBtn := minui.NewButton("opt_save", "Save")
	saveBtn.SetPosition(modalW/2-110, startY+5*rowGap+8)
	saveBtn.SetSize(100, 32)
	m.AddChild(saveBtn)

	closeBtn := minui.NewButton("opt_close", "Close")
	closeBtn.SetPosition(modalW/2+10, startY+5*rowGap+8)
	closeBtn.SetSize(100, 32)
	m.AddChild(closeBtn)

	om := &OptionsModal{
		modal:            m,
		crtInput:         crtInput,
		pressDelayInput:  pressDelayInput,
		renderScaleInput: renderScaleInput,
		npcDelayInput:    npcDelayInput,
		fullscreenToggle: fsToggle,
		saveBtn:          saveBtn,
	}

	saveBtn.OnClick = func() { om.apply() }
	closeBtn.OnClick = func() { om.Close() }

	return om
}

func (om *OptionsModal) apply() {
	cfg := config.Global()

	if v, err := strconv.ParseFloat(om.crtInput.Text, 64); err == nil {
		cfg.CRTIntensity = v
	}
	if v, err := strconv.Atoi(om.pressDelayInput.Text); err == nil {
		cfg.PressDelay = v
	}
	if v, err := strconv.ParseFloat(om.renderScaleInput.Text, 64); err == nil {
		cfg.RenderScale = v
	}
	if v, err := strconv.Atoi(om.npcDelayInput.Text); err == nil {
		cfg.NpcTurnDelayTicks = v
	}
	cfg.Fullscreen = om.fullscreenToggle.On
	ebiten.SetFullscreen(cfg.Fullscreen)

	if err := config.Save(); err != nil {
		log.Printf("options save failed: %v", err)
		om.statusMsg = "Save failed."
	} else {
		om.statusMsg = "Saved."
	}
}

func (om *OptionsModal) Open() {
	cfg := config.Global()
	om.crtInput.Text = fmt.Sprintf("%.2f", cfg.CRTIntensity)
	om.pressDelayInput.Text = strconv.Itoa(cfg.PressDelay)
	om.renderScaleInput.Text = fmt.Sprintf("%.1f", cfg.RenderScale)
	om.npcDelayInput.Text = strconv.Itoa(cfg.NpcTurnDelayTicks)
	om.fullscreenToggle.On = cfg.Fullscreen
	om.statusMsg = ""
	om.modal.SetVisible(true)
	om.Visible = true
}

func (om *OptionsModal) Close() {
	om.modal.SetVisible(false)
	om.Visible = false
}

func (om *OptionsModal) Update() {
	if !om.Visible {
		return
	}
	om.modal.Update()
}

func (om *OptionsModal) Draw(screen *ebiten.Image) {
	if !om.Visible {
		return
	}
	om.modal.Draw(screen)

	cfg := config.Global()
	cx := cfg.ScreenWidth/2 - 360/2
	cy := cfg.ScreenHeight/2 - 368/2

	labelColor := color.RGBA{180, 200, 180, 255}
	const fontSize = 13
	const titleBarH = 30
	drawLabel := func(text string, row int) {
		mlge_text.Draw(screen, text, fontSize, cx+20, cy+titleBarH+50+row*44+8, labelColor)
	}
	drawLabel("CRT Intensity (0-1):", 0)
	drawLabel("Press Delay (ticks):", 1)
	drawLabel("Render Scale:", 2)
	drawLabel("NPC Turn Delay:", 3)
	drawLabel("Fullscreen:", 4)

	if om.statusMsg != "" {
		mlge_text.Draw(screen, om.statusMsg, fontSize, cx+20, cy+titleBarH+50+5*44+8, color.RGBA{120, 200, 120, 255})
	}
}
