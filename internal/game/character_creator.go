package game

import (
	"fmt"
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mechanical-lich/mlge/resource"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/ccconfig"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/lore"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

const (
	ccModalW = 760
	ccModalH = 560
)

// CharacterCreator is the full-screen character creation modal shown at game start.
// It collects name, stats, class, skills, and background, then fires OnComplete.
type CharacterCreator struct {
	modal *minui.Modal
	tabs  *minui.TabPanel

	// Name tab
	nameInput *minui.TextInput

	// Stats tab
	statLabels  [6]*minui.Label // PH, AG, MA, CL, LD, HTCS
	pointsLabel *minui.Label
	stats       [6]int // current values for PH, AG, MA, CL, LD, HTCS

	// Class tab
	classList     *minui.ListBox
	classDesc     *minui.ScrollingTextArea
	classIDs      []string
	selectedClass int

	// Skills tab
	skillList      *minui.ListBox
	skillDesc      *minui.ScrollingTextArea
	skillPoints    *minui.Label
	buySkillBtn    *minui.Button
	forgetSkillBtn *minui.Button
	skillIDs       []string
	chosenSkills   []string
	upgradePoints  int
	selectedSkill  int

	// Appearance tab
	bodyTypeList      *minui.ListBox
	skinList          *minui.ListBox
	hairList          *minui.ListBox
	appearancePreview *minui.ImageWidget
	bodyType          string // "mid" or "slim"
	bodyIndex         int    // 0-4
	hairIndex         int    // -1 = none, 0-4

	// Background tab
	bgList     *minui.ListBox
	bgDesc     *minui.ScrollingTextArea
	bgIDs      []string
	selectedBg int

	// OnComplete is called when the player clicks Start.
	OnComplete func(data CharacterData)
}

const (
	statPH   = 0
	statAG   = 1
	statMA   = 2
	statCL   = 3
	statLD   = 4
	statHTCS = 5
)

var statNames = [6]string{"PH", "AG", "MA", "CL", "LD", "HTCS"}

func NewCharacterCreator() *CharacterCreator {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - ccModalW) / 2
	my := (cfg.ScreenHeight - ccModalH) / 2

	cc := &CharacterCreator{
		stats:         [6]int{10, 10, 10, 10, 8, 30},
		upgradePoints: 1,
		selectedClass: -1,
		selectedSkill: -1,
		selectedBg:    -1,
		bodyType:      "mid",
		bodyIndex:     0,
		hairIndex:     0,
	}

	// Non-closable modal.
	cc.modal = minui.NewModal("cc_modal", "Character Creation", ccModalW, ccModalH)
	cc.modal.SetPosition(mx, my)
	cc.modal.Closeable = false

	// Tab panel fills most of the modal content area, leaving room for Start button.
	cc.tabs = minui.NewTabPanel("cc_tabs", ccModalW-20, ccModalH-110)
	cc.tabs.SetPosition(10, 10)
	cc.modal.AddChild(cc.tabs)

	cc.buildNameTab()
	cc.buildStatsTab()
	cc.buildClassTab()
	cc.buildSkillsTab()
	cc.buildBackgroundTab()
	cc.buildAppearanceTab()

	// Start button — bottom-right of modal, within content area.
	startBtn := minui.NewButton("cc_start", "Start")
	startBtn.SetPosition(ccModalW-140, ccModalH-70)
	startBtn.SetSize(120, 30)
	startBtn.OnClick = func() { cc.submit() }
	cc.modal.AddChild(startBtn)

	cc.refreshClassList()
	cc.refreshSkillList()
	cc.refreshBgList()

	return cc
}

// -----------------------------------------------------------------------
// Tab builders
// -----------------------------------------------------------------------

func (cc *CharacterCreator) buildNameTab() {
	panel := minui.NewPanel("cc_name_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	lbl := minui.NewLabel("cc_name_lbl", "Character Name:")
	lbl.SetPosition(20, 30)
	lbl.SetSize(180, 24)
	panel.AddChild(lbl)

	cc.nameInput = minui.NewTextInput("cc_name_input", "Enter name…")
	cc.nameInput.SetPosition(20, 60)
	cc.nameInput.SetSize(300, 30)
	panel.AddChild(cc.nameInput)

	randomBtn := minui.NewButton("cc_random_name", "Random")
	randomBtn.SetPosition(330, 60)
	randomBtn.SetSize(80, 30)
	randomBtn.OnClick = func() {
		cc.nameInput.Text = lore.RandomName("crew")
	}
	panel.AddChild(randomBtn)

	cc.tabs.AddTab("name", "Name", panel)
}

func (cc *CharacterCreator) buildStatsTab() {
	panel := minui.NewPanel("cc_stats_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	cc.pointsLabel = minui.NewLabel("cc_pts_lbl", "")
	cc.pointsLabel.SetPosition(20, 15)
	cc.pointsLabel.SetSize(300, 24)
	panel.AddChild(cc.pointsLabel)

	for i := 0; i < 6; i++ {
		idx := i
		y := 50 + i*50

		lbl := minui.NewLabel(fmt.Sprintf("cc_stat_lbl_%d", i), "")
		lbl.SetPosition(20, y+4)
		lbl.SetSize(100, 24)
		panel.AddChild(lbl)
		cc.statLabels[i] = lbl

		minusBtn := minui.NewButton(fmt.Sprintf("cc_stat_minus_%d", i), "-")
		minusBtn.SetPosition(130, y)
		minusBtn.SetSize(30, 30)
		minusBtn.OnClick = func() { cc.adjustStat(idx, -1) }
		panel.AddChild(minusBtn)

		plusBtn := minui.NewButton(fmt.Sprintf("cc_stat_plus_%d", i), "+")
		plusBtn.SetPosition(170, y)
		plusBtn.SetSize(30, 30)
		plusBtn.OnClick = func() { cc.adjustStat(idx, 1) }
		panel.AddChild(plusBtn)
	}

	cc.refreshStatsDisplay()
	cc.tabs.AddTab("stats", "Stats", panel)
}

func (cc *CharacterCreator) buildClassTab() {
	panel := minui.NewPanel("cc_class_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	cc.classList = minui.NewListBox("cc_class_list", []string{})
	cc.classList.SetPosition(10, 10)
	cc.classList.SetSize(200, 300)
	cc.classList.Layout()
	cc.classList.OnSelect = func(idx int, _ string) {
		cc.selectedClass = idx
		cc.refreshClassDesc()
		// Switching class resets skill choices.
		cc.chosenSkills = nil
		cc.upgradePoints = 1
		cc.refreshSkillList()
		cc.refreshSkillDesc()
		cc.refreshSkillPoints()
	}
	panel.AddChild(cc.classList)

	cc.classDesc = minui.NewScrollingTextArea("cc_class_desc", 480, 300)
	cc.classDesc.SetPosition(220, 10)
	cc.classDesc.LineHeight = 14
	panel.AddChild(cc.classDesc)

	cc.tabs.AddTab("class", "Class", panel)
}

func (cc *CharacterCreator) buildSkillsTab() {
	panel := minui.NewPanel("cc_skills_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	cc.skillList = minui.NewListBox("cc_skill_list", []string{})
	cc.skillList.SetPosition(10, 10)
	cc.skillList.SetSize(200, 270)
	cc.skillList.Layout()
	cc.skillList.OnSelect = func(idx int, _ string) {
		cc.selectedSkill = idx
		cc.refreshSkillDesc()
	}
	panel.AddChild(cc.skillList)

	cc.skillDesc = minui.NewScrollingTextArea("cc_skill_desc", 480, 270)
	cc.skillDesc.SetPosition(220, 10)
	cc.skillDesc.LineHeight = 14
	panel.AddChild(cc.skillDesc)

	cc.skillPoints = minui.NewLabel("cc_skill_pts", "")
	cc.skillPoints.SetPosition(10, 290)
	cc.skillPoints.SetSize(200, 24)
	panel.AddChild(cc.skillPoints)

	cc.buySkillBtn = minui.NewButton("cc_buy_skill", "Learn Skill")
	cc.buySkillBtn.SetPosition(220, 285)
	cc.buySkillBtn.SetSize(120, 30)
	cc.buySkillBtn.OnClick = func() { cc.buySkill() }
	panel.AddChild(cc.buySkillBtn)

	cc.forgetSkillBtn = minui.NewButton("cc_forget_skill", "Forget Skill")
	cc.forgetSkillBtn.SetPosition(350, 285)
	cc.forgetSkillBtn.SetSize(120, 30)
	cc.forgetSkillBtn.OnClick = func() { cc.forgetSkill() }
	panel.AddChild(cc.forgetSkillBtn)

	cc.tabs.AddTab("skills", "Skills", panel)
}

func (cc *CharacterCreator) buildBackgroundTab() {
	panel := minui.NewPanel("cc_bg_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	cc.bgList = minui.NewListBox("cc_bg_list", []string{})
	cc.bgList.SetPosition(10, 10)
	cc.bgList.SetSize(200, 300)
	cc.bgList.Layout()
	cc.bgList.OnSelect = func(idx int, _ string) {
		cc.selectedBg = idx
		cc.refreshBgDesc()
	}
	panel.AddChild(cc.bgList)

	cc.bgDesc = minui.NewScrollingTextArea("cc_bg_desc", 480, 300)
	cc.bgDesc.SetPosition(220, 10)
	cc.bgDesc.LineHeight = 14
	panel.AddChild(cc.bgDesc)

	cc.tabs.AddTab("background", "Background", panel)
}

func (cc *CharacterCreator) buildAppearanceTab() {
	panel := minui.NewPanel("cc_appearance_panel")
	panel.SetPosition(0, cc.tabs.TabHeight)
	panel.SetSize(ccModalW-20, ccModalH-110-cc.tabs.TabHeight)

	bodyTypeLbl := minui.NewLabel("cc_body_type_lbl", "Body Type")
	bodyTypeLbl.SetPosition(10, 10)
	bodyTypeLbl.SetSize(160, 20)
	panel.AddChild(bodyTypeLbl)

	cc.bodyTypeList = minui.NewListBox("cc_body_type_list", []string{"Mid", "Slim"})
	cc.bodyTypeList.SetPosition(10, 35)
	cc.bodyTypeList.SetSize(160, 60)
	cc.bodyTypeList.Layout()
	cc.bodyTypeList.SelectedIndex = 0
	cc.bodyTypeList.OnSelect = func(_ int, val string) {
		if val == "Slim" {
			cc.bodyType = "slim"
		} else {
			cc.bodyType = "mid"
		}
		cc.refreshAppearancePreview()
	}
	panel.AddChild(cc.bodyTypeList)

	skinLbl := minui.NewLabel("cc_skin_lbl", "Skin Tone")
	skinLbl.SetPosition(10, 110)
	skinLbl.SetSize(160, 20)
	panel.AddChild(skinLbl)

	cc.skinList = minui.NewListBox("cc_skin_list", []string{"1", "2", "3", "4", "5"})
	cc.skinList.SetPosition(10, 135)
	cc.skinList.SetSize(160, 120)
	cc.skinList.Layout()
	cc.skinList.SelectedIndex = 0
	cc.skinList.OnSelect = func(idx int, _ string) {
		cc.bodyIndex = idx
		cc.refreshAppearancePreview()
	}
	panel.AddChild(cc.skinList)

	hairLbl := minui.NewLabel("cc_hair_lbl", "Hair Style")
	hairLbl.SetPosition(10, 270)
	hairLbl.SetSize(160, 20)
	panel.AddChild(hairLbl)

	cc.hairList = minui.NewListBox("cc_hair_list", []string{"Style 1", "Style 2", "Style 3", "Style 4", "Style 5", "Style 6", "Style 7", "Style 8", "Style 9", "Style 10", "None"})
	cc.hairList.SetPosition(10, 295)
	cc.hairList.SetSize(160, 130)
	cc.hairList.Layout()
	cc.hairList.SelectedIndex = 0
	cc.hairList.OnSelect = func(idx int, _ string) {
		if idx >= 10 {
			cc.hairIndex = -1
		} else {
			cc.hairIndex = idx
		}
		cc.refreshAppearancePreview()
	}
	panel.AddChild(cc.hairList)

	// Preview widget — right half of the panel, scaled up 4x (32x48 → 128x192).
	const previewScale = 4
	cfg := config.Global()
	previewW := cfg.TileSizeW * previewScale
	previewH := cfg.TileSizeH * previewScale
	previewX := 220 + (ccModalW-20-220-previewW)/2
	previewY := (ccModalH - 110 - cc.tabs.TabHeight - previewH) / 2
	cc.appearancePreview = minui.NewImageWidget("cc_appearance_preview", previewW, previewH)
	cc.appearancePreview.SetPosition(previewX, previewY)
	panel.AddChild(cc.appearancePreview)

	cc.tabs.AddTab("appearance", "Appearance", panel)
	cc.refreshAppearancePreview()
}

// refreshAppearancePreview redraws the composited character preview image.
func (cc *CharacterCreator) refreshAppearancePreview() {
	if cc.appearancePreview == nil {
		return
	}
	cfg := config.Global()
	spW := cfg.TileSizeW
	spH := cfg.TileSizeH
	bt := cc.bodyType

	out := ebiten.NewImage(spW, spH)

	drawLayer := func(texName string, idx int) {
		tex, ok := resource.Textures[texName]
		if !ok {
			return
		}
		srcX := idx * spW
		srcRect := image.Rect(srcX, 0, srcX+spW, spH)
		op := &ebiten.DrawImageOptions{}
		out.DrawImage(tex.SubImage(srcRect).(*ebiten.Image), op)
	}

	drawLayer(bt+"_body", cc.bodyIndex)
	if cc.hairIndex >= 0 {
		drawLayer(bt+"_hair", cc.hairIndex)
	}

	cc.appearancePreview.Image = out
}

// -----------------------------------------------------------------------
// Refresh helpers
// -----------------------------------------------------------------------

func (cc *CharacterCreator) remainingPoints() int {
	total := 10
	for _, v := range cc.stats {
		total -= v - 10
	}
	return total
}

func (cc *CharacterCreator) refreshStatsDisplay() {
	remaining := cc.remainingPoints()
	cc.pointsLabel.Text = fmt.Sprintf("Points remaining: %d", remaining)
	for i, v := range cc.stats {
		cc.statLabels[i].Text = fmt.Sprintf("%s: %d", statNames[i], v)
	}
}

func (cc *CharacterCreator) adjustStat(idx, delta int) {
	if delta > 0 && cc.remainingPoints() <= 0 {
		return
	}
	cc.stats[idx] += delta
	cc.refreshStatsDisplay()
}

func (cc *CharacterCreator) refreshClassList() {
	defs := class.All()
	cc.classIDs = make([]string, len(defs))
	names := make([]string, len(defs))
	for i, d := range defs {
		cc.classIDs[i] = d.ID
		names[i] = d.Name
	}
	cc.classList.SetItems(names)
	cc.refreshClassDesc()
}

func (cc *CharacterCreator) refreshClassDesc() {
	cc.classDesc.Clear()
	if cc.selectedClass < 0 || cc.selectedClass >= len(cc.classIDs) {
		cc.classDesc.AddText("Select a class to see its description.")
		return
	}
	def := class.Get(cc.classIDs[cc.selectedClass])
	if def == nil {
		return
	}
	cc.classDesc.AddText(def.Name)
	cc.classDesc.AddText("")
	for _, line := range mlge_text.Wrap(def.Description, 40, 0) {
		cc.classDesc.AddText(line)
	}

	if len(def.StartingSkills) > 0 {
		cc.classDesc.AddText("")
		cc.classDesc.AddText("Starting Skills:")
		for _, sID := range def.StartingSkills {
			name := sID
			desc := ""
			if sd := skill.Get(sID); sd != nil {
				name = sd.Name
				desc = sd.Description
			}
			cc.classDesc.AddText("  • " + name)
			if desc != "" {
				for _, line := range mlge_text.Wrap(desc, 38, 4) {
					cc.classDesc.AddText(line)
				}
			}
		}
	}

	if len(def.Skills) > 0 {
		cc.classDesc.AddText("")
		cc.classDesc.AddText("Upgradeable Skills:")
		for _, sID := range def.Skills {
			name := sID
			desc := ""
			if sd := skill.Get(sID); sd != nil {
				name = sd.Name
				desc = sd.Description
			}
			cc.classDesc.AddText("  • " + name)
			if desc != "" {
				for _, line := range mlge_text.Wrap(desc, 38, 4) {
					cc.classDesc.AddText(line)
				}
			}
		}
	}
}

func (cc *CharacterCreator) refreshSkillList() {
	baseSkills := ccconfig.Get().BaseSkills

	var classSkills []string
	if cc.selectedClass >= 0 && cc.selectedClass < len(cc.classIDs) {
		if def := class.Get(cc.classIDs[cc.selectedClass]); def != nil {
			classSkills = def.Skills
		}
	}

	cc.skillIDs = make([]string, 0, len(baseSkills)+len(classSkills))
	names := make([]string, 0, len(baseSkills)+len(classSkills))

	for _, sID := range baseSkills {
		cc.skillIDs = append(cc.skillIDs, sID)
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		if slices.Contains(cc.chosenSkills, sID) {
			name = "✓ " + name
		}
		names = append(names, name)
	}

	for _, sID := range classSkills {
		cc.skillIDs = append(cc.skillIDs, sID)
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		if slices.Contains(cc.chosenSkills, sID) {
			name = "✓ " + name
		}
		names = append(names, name)
	}

	cc.skillList.SetItems(names)
	if cc.selectedSkill >= len(names) {
		cc.selectedSkill = -1
		cc.skillList.SelectedIndex = -1
	}
}

func (cc *CharacterCreator) refreshSkillDesc() {
	cc.skillDesc.Clear()
	if cc.selectedSkill < 0 || cc.selectedSkill >= len(cc.skillIDs) {
		cc.skillDesc.AddText("Select a skill to see its description.")
		return
	}
	sID := cc.skillIDs[cc.selectedSkill]
	sd := skill.Get(sID)
	if sd == nil {
		return
	}
	cc.skillDesc.AddText(sd.Name)
	cc.skillDesc.AddText("")
	for _, line := range mlge_text.Wrap(sd.Description, 40, 0) {
		cc.skillDesc.AddText(line)
	}
	if slices.Contains(cc.chosenSkills, sID) {
		cc.skillDesc.AddText("")
		cc.skillDesc.AddText("(already chosen)")
	}
}

func (cc *CharacterCreator) refreshSkillPoints() {
	cc.skillPoints.Text = fmt.Sprintf("Upgrade Points: %d", cc.upgradePoints)
}

func (cc *CharacterCreator) buySkill() {
	if cc.selectedSkill < 0 || cc.selectedSkill >= len(cc.skillIDs) {
		return
	}
	sID := cc.skillIDs[cc.selectedSkill]
	if slices.Contains(cc.chosenSkills, sID) {
		return
	}
	if cc.upgradePoints <= 0 {
		return
	}
	cc.chosenSkills = append(cc.chosenSkills, sID)
	cc.upgradePoints--
	cc.refreshSkillList()
	cc.refreshSkillDesc()
	cc.refreshSkillPoints()
}

func (cc *CharacterCreator) forgetSkill() {
	if cc.selectedSkill < 0 || cc.selectedSkill >= len(cc.skillIDs) {
		return
	}
	sID := cc.skillIDs[cc.selectedSkill]
	idx := slices.Index(cc.chosenSkills, sID)
	if idx < 0 {
		return
	}
	cc.chosenSkills = slices.Delete(cc.chosenSkills, idx, idx+1)
	cc.upgradePoints++
	cc.refreshSkillList()
	cc.refreshSkillDesc()
	cc.refreshSkillPoints()
}

func (cc *CharacterCreator) refreshBgList() {
	defs := background.All()
	cc.bgIDs = make([]string, len(defs))
	names := make([]string, len(defs))
	for i, d := range defs {
		cc.bgIDs[i] = d.ID
		names[i] = d.Name
	}
	cc.bgList.SetItems(names)
	cc.refreshBgDesc()
}

func (cc *CharacterCreator) refreshBgDesc() {
	cc.bgDesc.Clear()
	if cc.selectedBg < 0 || cc.selectedBg >= len(cc.bgIDs) {
		cc.bgDesc.AddText("Select a background to see its description.")
		return
	}
	def := background.Get(cc.bgIDs[cc.selectedBg])
	if def == nil {
		return
	}
	cc.bgDesc.AddText(def.Name)
	cc.bgDesc.AddText("")
	for _, line := range mlge_text.Wrap(def.Description, 40, 0) {
		cc.bgDesc.AddText(line)
	}
	cc.bgDesc.AddText("")
	cc.bgDesc.AddText("Grants:")
	for _, sID := range def.Skills {
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		cc.bgDesc.AddText("  • " + name)
	}
}

// -----------------------------------------------------------------------
// Submission
// -----------------------------------------------------------------------

func (cc *CharacterCreator) submit() {
	name := cc.nameInput.Text
	if name == "" {
		name = lore.RandomName("crew")
	}
	classID := ""
	if cc.selectedClass >= 0 && cc.selectedClass < len(cc.classIDs) {
		classID = cc.classIDs[cc.selectedClass]
	}
	bgID := ""
	if cc.selectedBg >= 0 && cc.selectedBg < len(cc.bgIDs) {
		bgID = cc.bgIDs[cc.selectedBg]
	}
	if cc.OnComplete != nil {
		cc.OnComplete(CharacterData{
			Name:         name,
			PH:           cc.stats[statPH],
			AG:           cc.stats[statAG],
			MA:           cc.stats[statMA],
			CL:           cc.stats[statCL],
			LD:           cc.stats[statLD],
			HTCS:         cc.stats[statHTCS],
			ClassID:      classID,
			ChosenSkills: cc.chosenSkills,
			BackgroundID: bgID,
			BodyType:     cc.bodyType,
			BodyIndex:    cc.bodyIndex,
			HairIndex:    cc.hairIndex,
		})
	}
}

// -----------------------------------------------------------------------
// Update / Draw
// -----------------------------------------------------------------------

func (cc *CharacterCreator) Update() {
	cc.modal.Update()
	cc.refreshSkillPoints()
}

func (cc *CharacterCreator) Draw(screen *ebiten.Image) {
	cc.modal.Draw(screen)
}
