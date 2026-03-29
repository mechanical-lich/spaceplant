package game

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/config"
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
	statLabels    [4]*minui.Label // Str, Dex, Int, Wis
	pointsLabel   *minui.Label
	stats         [4]int // current values for Str, Dex, Int, Wis

	// Class tab
	classList    *minui.ListBox
	classDesc    *minui.ScrollingTextArea
	classIDs     []string
	selectedClass int

	// Skills tab
	skillList     *minui.ListBox
	skillDesc     *minui.ScrollingTextArea
	skillPoints   *minui.Label
	buySkillBtn   *minui.Button
	skillIDs      []string
	chosenSkills  []string
	upgradePoints int
	selectedSkill int

	// Background tab
	bgList    *minui.ListBox
	bgDesc    *minui.ScrollingTextArea
	bgIDs     []string
	selectedBg int

	// OnComplete is called when the player clicks Start.
	OnComplete func(data CharacterData)
}

const (
	statStr = 0
	statDex = 1
	statInt = 2
	statWis = 3
)

var statNames = [4]string{"Str", "Dex", "Int", "Wis"}

func NewCharacterCreator() *CharacterCreator {
	cfg := config.Global()
	mx := (cfg.ScreenWidth - ccModalW) / 2
	my := (cfg.ScreenHeight - ccModalH) / 2

	cc := &CharacterCreator{
		stats:         [4]int{10, 10, 10, 10},
		upgradePoints: 1,
		selectedClass: -1,
		selectedSkill: -1,
		selectedBg:    -1,
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

	// Start button — bottom-right of modal, within content area.
	startBtn := minui.NewButton("cc_start", "Start")
	startBtn.SetPosition(ccModalW-140, ccModalH-70)
	startBtn.SetSize(120, 30)
	startBtn.OnClick = func() { cc.submit() }
	cc.modal.AddChild(startBtn)

	cc.refreshClassList()
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

	for i := 0; i < 4; i++ {
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
	cc.classDesc.AddText("")
	cc.classDesc.AddText("Skills:")
	for _, sID := range def.Skills {
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		cc.classDesc.AddText("  • " + name)
	}
}

func (cc *CharacterCreator) refreshSkillList() {
	if cc.selectedClass < 0 || cc.selectedClass >= len(cc.classIDs) {
		cc.skillList.SetItems([]string{})
		cc.skillIDs = nil
		return
	}
	def := class.Get(cc.classIDs[cc.selectedClass])
	if def == nil {
		cc.skillList.SetItems([]string{})
		cc.skillIDs = nil
		return
	}
	cc.skillIDs = make([]string, len(def.Skills))
	names := make([]string, len(def.Skills))
	for i, sID := range def.Skills {
		cc.skillIDs[i] = sID
		name := sID
		if sd := skill.Get(sID); sd != nil {
			name = sd.Name
		}
		if slices.Contains(cc.chosenSkills, sID) {
			name = "✓ " + name
		}
		names[i] = name
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
	sd := skill.Get(cc.skillIDs[cc.selectedSkill])
	if sd == nil {
		return
	}
	cc.skillDesc.AddText(sd.Name)
	cc.skillDesc.AddText("")
	for _, line := range mlge_text.Wrap(sd.Description, 40, 0) {
		cc.skillDesc.AddText(line)
	}
	if slices.Contains(cc.chosenSkills, cc.skillIDs[cc.selectedSkill]) {
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
		name = "Unnamed"
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
			Str:          cc.stats[statStr],
			Dex:          cc.stats[statDex],
			Int:          cc.stats[statInt],
			Wis:          cc.stats[statWis],
			ClassID:      classID,
			ChosenSkills: cc.chosenSkills,
			BackgroundID: bgID,
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
