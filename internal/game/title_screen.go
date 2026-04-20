package game

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/client"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/buildinfo"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/lore"
	"github.com/mechanical-lich/spaceplant/internal/scenario"
	"github.com/mechanical-lich/spaceplant/internal/stationconfig"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
)

const savesDir = "saves"

type titleScreen int

const (
	screenMain           titleScreen = iota
	screenStationBrowser             // choose or generate a station
	screenPlayerBrowser              // choose a player run on the selected station
	screenNewStationName             // enter name for a new station
)

// TitleScreenState is the initial client state shown before starting or loading a game.
// It implements client.ClientState.
type TitleScreenState struct {
	sim       *SimWorld
	simState  *MainSimState
	transport transport.ClientTransport

	screen titleScreen

	// Main menu buttons
	newStationBtn     *minui.Button
	browseStationsBtn *minui.Button
	optionsBtn        *minui.Button
	quitBtn           *minui.Button

	// Station browser
	stationList         *minui.ListBox
	stations            []StationMeta
	selectStationBtn    *minui.Button
	backFromStationsBtn *minui.Button

	// Player browser
	playerList         *minui.ListBox
	players            []PlayerRunMeta
	selectedStation    StationMeta
	newPlayerBtn       *minui.Button
	continueBtn        *minui.Button
	backFromPlayersBtn *minui.Button

	// New station name input
	stationNameInput  *minui.TextInput
	randomNameBtn     *minui.Button
	confirmNameBtn    *minui.Button
	cancelNameBtn     *minui.Button
	scenarioPicker    *minui.SelectBox
	scenarioPickerIDs []string // parallel to scenarioPicker items; "" means random

	// Station config spinners
	cfgCrewCapacity        int
	cfgScienceLabCount     int
	cfgMedCount            int
	cfgEngineeringCapacity int
	cfgSecurityCapacity    int
	cfgLifePodBayCount     int
	cfgSelfDestructEnabled bool
	cfgCrewMinus           *minui.Button
	cfgCrewPlus            *minui.Button
	cfgSciMinus            *minui.Button
	cfgSciPlus             *minui.Button
	cfgMedMinus            *minui.Button
	cfgMedPlus             *minui.Button
	cfgEngMinus            *minui.Button
	cfgEngPlus             *minui.Button
	cfgSecMinus            *minui.Button
	cfgSecPlus             *minui.Button
	cfgPodMinus            *minui.Button
	cfgPodPlus             *minui.Button
	cfgSelfDestructToggle  *minui.Toggle

	optionsModal *OptionsModal

	errMsg string
	done   bool
	next   client.ClientState
}

var _ client.ClientState = (*TitleScreenState)(nil)

// NewTitleScreenState creates the title screen.
func NewTitleScreenState(sim *SimWorld, simState *MainSimState, t transport.ClientTransport) *TitleScreenState {
	ts := &TitleScreenState{
		sim:       sim,
		simState:  simState,
		transport: t,
		screen:    screenMain,
	}
	ts.buildMainMenu()
	ts.optionsModal = newOptionsModal()
	return ts
}

func (ts *TitleScreenState) buildMainMenu() {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2
	btnW := 220
	btnH := 36
	btnX := cx - btnW/2

	ts.newStationBtn = minui.NewButton("title_new_station", "New Station")
	ts.newStationBtn.SetPosition(btnX, 340)
	ts.newStationBtn.SetSize(btnW, btnH)
	ts.newStationBtn.OnClick = func() { ts.screen = screenNewStationName }

	ts.browseStationsBtn = minui.NewButton("title_browse", "Browse Stations")
	ts.browseStationsBtn.SetPosition(btnX, 340+btnH+12)
	ts.browseStationsBtn.SetSize(btnW, btnH)
	ts.browseStationsBtn.OnClick = func() { ts.openStationBrowser() }

	ts.optionsBtn = minui.NewButton("title_options", "Options")
	ts.optionsBtn.SetPosition(btnX, 340+2*(btnH+12))
	ts.optionsBtn.SetSize(btnW, btnH)
	ts.optionsBtn.OnClick = func() { ts.optionsModal.Open() }

	ts.quitBtn = minui.NewButton("title_quit", "Quit")
	ts.quitBtn.SetPosition(btnX, 340+3*(btnH+12))
	ts.quitBtn.SetSize(btnW, btnH)
	ts.quitBtn.OnClick = func() { os.Exit(0) }

	// New station name input
	ts.stationNameInput = minui.NewTextInput("station_name_input", "")
	ts.stationNameInput.SetPosition(cx-150, 315)
	ts.stationNameInput.SetSize(220, 28)

	randomNameBtn := minui.NewButton("random_station_name", "Random")
	randomNameBtn.SetPosition(cx+80, 315)
	randomNameBtn.SetSize(80, 28)
	randomNameBtn.OnClick = func() {
		ts.stationNameInput.Text = lore.RandomStationName()
	}
	ts.randomNameBtn = randomNameBtn

	// Scenario picker: "Random" + one entry per enabled scenario.
	scenarios := scenario.AllEnabled()
	pickerLabels := make([]string, 0, len(scenarios)+1)
	ts.scenarioPickerIDs = make([]string, 0, len(scenarios)+1)
	pickerLabels = append(pickerLabels, "Random")
	ts.scenarioPickerIDs = append(ts.scenarioPickerIDs, "")
	for _, s := range scenarios {
		pickerLabels = append(pickerLabels, s.Name)
		ts.scenarioPickerIDs = append(ts.scenarioPickerIDs, s.ID)
	}
	ts.scenarioPicker = minui.NewSelectBox("scenario_picker", pickerLabels)
	ts.scenarioPicker.SetPosition(cx-150, 370)
	ts.scenarioPicker.SetSize(300, 28)
	ts.scenarioPicker.SelectByIndex(0)

	// Station config spinners — seed from defaults.
	def := stationconfig.Get()
	ts.cfgCrewCapacity = def.CrewCapacity
	ts.cfgScienceLabCount = def.ScienceLabCount
	ts.cfgMedCount = def.MedCount
	ts.cfgEngineeringCapacity = def.EngineeringCapacity
	ts.cfgSecurityCapacity = def.SecurityCapacity
	ts.cfgLifePodBayCount = def.LifePodBayCount
	ts.cfgSelfDestructEnabled = def.SelfDestructEnabled

	ts.cfgCrewMinus, ts.cfgCrewPlus = ts.makeSpinner("cfg_crew", cx+60, 416, func(d int) { ts.cfgCrewCapacity = clampSpin(ts.cfgCrewCapacity+d, 1, 100) })
	ts.cfgSciMinus, ts.cfgSciPlus = ts.makeSpinner("cfg_sci", cx+60, 441, func(d int) { ts.cfgScienceLabCount = clampSpin(ts.cfgScienceLabCount+d, 0, 30) })
	ts.cfgMedMinus, ts.cfgMedPlus = ts.makeSpinner("cfg_med", cx+60, 466, func(d int) { ts.cfgMedCount = clampSpin(ts.cfgMedCount+d, 0, 30) })
	ts.cfgEngMinus, ts.cfgEngPlus = ts.makeSpinner("cfg_eng", cx+60, 491, func(d int) { ts.cfgEngineeringCapacity = clampSpin(ts.cfgEngineeringCapacity+d, 0, 30) })
	ts.cfgSecMinus, ts.cfgSecPlus = ts.makeSpinner("cfg_sec", cx+60, 516, func(d int) { ts.cfgSecurityCapacity = clampSpin(ts.cfgSecurityCapacity+d, 0, 30) })
	ts.cfgPodMinus, ts.cfgPodPlus = ts.makeSpinner("cfg_pod", cx+60, 541, func(d int) { ts.cfgLifePodBayCount = clampSpin(ts.cfgLifePodBayCount+d, 0, 20) })

	ts.cfgSelfDestructToggle = minui.NewToggle("cfg_selfdestruct", "Self-Destruct")
	ts.cfgSelfDestructToggle.On = ts.cfgSelfDestructEnabled
	ts.cfgSelfDestructToggle.SetPosition(cx-150, 566)
	ts.cfgSelfDestructToggle.SetSize(200, 24)
	ts.cfgSelfDestructToggle.OnChange = func(on bool) { ts.cfgSelfDestructEnabled = on }

	ts.confirmNameBtn = minui.NewButton("confirm_name", "Generate")
	ts.confirmNameBtn.SetPosition(cx-80, 603)
	ts.confirmNameBtn.SetSize(160, 36)
	ts.confirmNameBtn.OnClick = func() { ts.generateNewStation() }

	ts.cancelNameBtn = minui.NewButton("cancel_name", "Cancel")
	ts.cancelNameBtn.SetPosition(cx-80, 648)
	ts.cancelNameBtn.SetSize(160, 36)
	ts.cancelNameBtn.OnClick = func() { ts.screen = screenMain }
}

// makeSpinner creates a minus and plus button for a numeric spinner at (x, y).
func (ts *TitleScreenState) makeSpinner(id string, x, y int, onChange func(delta int)) (*minui.Button, *minui.Button) {
	minus := minui.NewButton(id+"_minus", "-")
	minus.SetPosition(x, y)
	minus.SetSize(24, 24)
	minus.OnClick = func() { onChange(-1) }

	plus := minui.NewButton(id+"_plus", "+")
	plus.SetPosition(x+50, y)
	plus.SetSize(24, 24)
	plus.OnClick = func() { onChange(1) }

	return minus, plus
}

func clampSpin(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (ts *TitleScreenState) openStationBrowser() {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2

	stations, err := ListStations(savesDir)
	if err != nil {
		ts.errMsg = "Load failed: " + err.Error()
		return
	}
	ts.stations = stations
	ts.errMsg = ""

	labels := make([]string, len(stations))
	for i, s := range stations {
		labels[i] = s.Name
	}

	ts.stationList = minui.NewListBox("station_list", labels)
	ts.stationList.SetPosition(cx-200, 280)
	ts.stationList.SetSize(400, 300)
	ts.stationList.Layout()
	ts.stationList.OnSelect = func(idx int, _ string) {
		_ = idx
	}

	btnW := 160
	ts.selectStationBtn = minui.NewButton("select_station", "Select Station")
	ts.selectStationBtn.SetPosition(cx-btnW/2, 600)
	ts.selectStationBtn.SetSize(btnW, 34)
	ts.selectStationBtn.OnClick = func() { ts.openPlayerBrowser() }

	ts.backFromStationsBtn = minui.NewButton("back_stations", "Back")
	ts.backFromStationsBtn.SetPosition(cx-btnW/2, 644)
	ts.backFromStationsBtn.SetSize(btnW, 34)
	ts.backFromStationsBtn.OnClick = func() {
		ts.errMsg = ""
		ts.screen = screenMain
	}

	ts.screen = screenStationBrowser
}

func (ts *TitleScreenState) openPlayerBrowser() {
	if ts.stationList == nil {
		return
	}
	idx := ts.stationList.SelectedIndex
	if idx < 0 || idx >= len(ts.stations) {
		ts.errMsg = "Select a station first."
		return
	}
	ts.selectedStation = ts.stations[idx]
	ts.errMsg = ""

	cfg := config.Global()
	cx := cfg.ScreenWidth / 2

	players, err := ListPlayerRuns(savesDir, ts.selectedStation.StationID)
	if err != nil {
		ts.errMsg = "Load failed: " + err.Error()
		return
	}
	ts.players = players

	labels := make([]string, len(players))
	for i, p := range players {
		label := p.Name
		if p.ClassName != "" {
			label += " — " + p.ClassName
		}
		label += fmt.Sprintf(" — Floor %d", p.CurrentZ+1)
		if p.Dead {
			label += " (dead)"
		}
		labels[i] = label
	}

	ts.playerList = minui.NewListBox("player_list", labels)
	ts.playerList.SetPosition(cx-200, 280)
	ts.playerList.SetSize(400, 300)
	ts.playerList.Layout()

	btnW := 160
	ts.newPlayerBtn = minui.NewButton("new_player", "New Player")
	ts.newPlayerBtn.SetPosition(cx-btnW/2, 600)
	ts.newPlayerBtn.SetSize(btnW, 34)
	ts.newPlayerBtn.OnClick = func() { ts.startNewPlayerOnStation() }

	ts.continueBtn = minui.NewButton("continue_run", "Continue")
	ts.continueBtn.SetPosition(cx-btnW/2, 644)
	ts.continueBtn.SetSize(btnW, 34)
	ts.continueBtn.OnClick = func() { ts.continuePlayerRun() }

	ts.backFromPlayersBtn = minui.NewButton("back_players", "Back")
	ts.backFromPlayersBtn.SetPosition(cx-btnW/2, 688)
	ts.backFromPlayersBtn.SetSize(btnW, 34)
	ts.backFromPlayersBtn.OnClick = func() {
		ts.errMsg = ""
		ts.openStationBrowser()
	}

	ts.screen = screenPlayerBrowser
}

func (ts *TitleScreenState) generateNewStation() {
	name := ts.stationNameInput.Text
	if name == "" {
		name = fmt.Sprintf("Station %s", generateID()[:4])
	}
	idx := ts.scenarioPicker.SelectedIndex
	if idx > 0 && idx < len(ts.scenarioPickerIDs) {
		if err := scenario.SelectByID(ts.scenarioPickerIDs[idx]); err != nil {
			ts.errMsg = err.Error()
			return
		}
	} else {
		if err := scenario.SelectRandom(); err != nil {
			ts.errMsg = err.Error()
			return
		}
	}
	wincondition.LoadFromRules(scenario.Active().WinConditions)
	stationconfig.Set(stationconfig.Config{
		CrewCapacity:        ts.cfgCrewCapacity,
		ScienceLabCount:     ts.cfgScienceLabCount,
		MedCount:            ts.cfgMedCount,
		EngineeringCapacity: ts.cfgEngineeringCapacity,
		SecurityCapacity:    ts.cfgSecurityCapacity,
		LifePodBayCount:     ts.cfgLifePodBayCount,
		SelfDestructEnabled: ts.cfgSelfDestructEnabled,
	})
	if err := ts.sim.RegenerateLevel(); err != nil {
		ts.errMsg = "Generate failed: " + err.Error()
		return
	}
	ts.sim.StationName = name
	ts.simState.ResetPhase()
	// Save the station immediately so it appears in the list.
	if err := SaveStation(ts.sim, savesDir); err != nil {
		ts.errMsg = "Save failed: " + err.Error()
		return
	}
	ts.next = NewSPClientState(ts.sim, ts.simState, ts.transport)
	ts.done = true
}

func (ts *TitleScreenState) startNewPlayerOnStation() {
	if err := LoadStationIntoSimWorld(ts.sim, ts.selectedStation.StationID, savesDir); err != nil {
		ts.errMsg = "Load failed: " + err.Error()
		return
	}
	ts.simState.ResetPhase()
	ts.next = NewSPClientState(ts.sim, ts.simState, ts.transport)
	ts.done = true
}

func (ts *TitleScreenState) continuePlayerRun() {
	if ts.playerList == nil {
		return
	}
	idx := ts.playerList.SelectedIndex
	if idx < 0 || idx >= len(ts.players) {
		ts.errMsg = "Select a player run first."
		return
	}
	p := ts.players[idx]
	if p.Dead {
		ts.errMsg = "That player is dead. Start a new player instead."
		return
	}
	if err := LoadFullGame(ts.sim, p.StationID, p.PlayerRunID, savesDir); err != nil {
		ts.errMsg = "Load failed: " + err.Error()
		return
	}
	ts.simState.ResetPhase()
	ts.next = NewSPClientState(ts.sim, ts.simState, ts.transport)
	ts.done = true
}

// Update implements client.ClientState.
func (ts *TitleScreenState) Update(_ *transport.Snapshot) client.ClientState {
	switch ts.screen {
	case screenMain:
		ts.updateMain()
	case screenNewStationName:
		ts.updateNewStationName()
	case screenStationBrowser:
		ts.updateStationBrowser()
	case screenPlayerBrowser:
		ts.updatePlayerBrowser()
	}
	return ts.next
}

func (ts *TitleScreenState) updateMain() {
	if ts.optionsModal.Visible {
		ts.optionsModal.Update()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	ts.newStationBtn.Update()
	ts.browseStationsBtn.Update()
	ts.optionsBtn.Update()
	ts.quitBtn.Update()
}

func (ts *TitleScreenState) updateNewStationName() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ts.screen = screenMain
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ts.generateNewStation()
		return
	}
	ts.stationNameInput.Update()
	ts.randomNameBtn.Update()
	ts.scenarioPicker.Update()
	ts.cfgCrewMinus.Update()
	ts.cfgCrewPlus.Update()
	ts.cfgSciMinus.Update()
	ts.cfgSciPlus.Update()
	ts.cfgMedMinus.Update()
	ts.cfgMedPlus.Update()
	ts.cfgEngMinus.Update()
	ts.cfgEngPlus.Update()
	ts.cfgSecMinus.Update()
	ts.cfgSecPlus.Update()
	ts.cfgPodMinus.Update()
	ts.cfgPodPlus.Update()
	ts.cfgSelfDestructToggle.Update()
	ts.confirmNameBtn.Update()
	ts.cancelNameBtn.Update()
}

func (ts *TitleScreenState) updateStationBrowser() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ts.screen = screenMain
		return
	}
	ts.stationList.Update()
	ts.selectStationBtn.Update()
	ts.backFromStationsBtn.Update()
}

func (ts *TitleScreenState) updatePlayerBrowser() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ts.openStationBrowser()
		return
	}
	ts.playerList.Update()
	ts.newPlayerBtn.Update()
	ts.continueBtn.Update()
	ts.backFromPlayersBtn.Update()
}

// Draw implements client.ClientState.
func (ts *TitleScreenState) Draw(screen *ebiten.Image) {
	cfg := config.Global()
	w := float64(cfg.ScreenWidth)
	h := float64(cfg.ScreenHeight)

	ebitenutil.DrawRect(screen, 0, 0, w, h, color.RGBA{10, 10, 18, 255})

	titleText := "Space Plants!"
	mlge_text.Draw(screen, titleText, 52, cfg.ScreenWidth/2-len(titleText)*52*3/10/2, 180, color.RGBA{180, 230, 120, 255})

	subText := "A Sci-Fi Roguelike"
	mlge_text.Draw(screen, subText, 18, cfg.ScreenWidth/2-len(subText)*18*3/10/2, 250, color.RGBA{130, 160, 100, 255})

	switch ts.screen {
	case screenMain:
		ts.drawMain(screen)
	case screenNewStationName:
		ts.drawNewStationName(screen)
	case screenStationBrowser:
		ts.drawStationBrowser(screen)
	case screenPlayerBrowser:
		ts.drawPlayerBrowser(screen)
	}

	if ts.errMsg != "" {
		mlge_text.Draw(screen, ts.errMsg, 13, cfg.ScreenWidth/2-200, int(h)-60, color.RGBA{220, 80, 80, 255})
	}

	ver := buildinfo.Version
	mlge_text.Draw(screen, ver, 11, cfg.ScreenWidth-len(ver)*11*4/10-8, int(h)-18, color.RGBA{80, 90, 80, 255})
}

func (ts *TitleScreenState) drawMain(screen *ebiten.Image) {
	ts.newStationBtn.Draw(screen)
	ts.browseStationsBtn.Draw(screen)
	ts.optionsBtn.Draw(screen)
	ts.quitBtn.Draw(screen)

	cfg := config.Global()
	h := float64(cfg.ScreenHeight)
	hint := "[N] New Station    [B] Browse    [Q] Quit"
	mlge_text.Draw(screen, hint, 12, cfg.ScreenWidth/2-len(hint)*12*3/10/2, int(h)-40, color.RGBA{90, 100, 90, 255})

	ts.optionsModal.Draw(screen)
}

func (ts *TitleScreenState) drawNewStationName(screen *ebiten.Image) {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2
	labelColor := color.RGBA{180, 200, 180, 255}
	dimColor := color.RGBA{130, 150, 130, 255}

	mlge_text.Draw(screen, "Station Name:", 14, cx-150, 298, labelColor)
	ts.stationNameInput.Draw(screen)
	ts.randomNameBtn.Draw(screen)

	mlge_text.Draw(screen, "Scenario:", 14, cx-150, 354, labelColor)
	ts.scenarioPicker.Draw(screen)

	mlge_text.Draw(screen, "Station Configuration", 13, cx-150, 402, dimColor)

	type spinRow struct {
		label string
		y     int
		val   int
		minus *minui.Button
		plus  *minui.Button
	}
	rows := []spinRow{
		{"Crew Capacity:", 416, ts.cfgCrewCapacity, ts.cfgCrewMinus, ts.cfgCrewPlus},
		{"Science Labs:", 441, ts.cfgScienceLabCount, ts.cfgSciMinus, ts.cfgSciPlus},
		{"Med Bays:", 466, ts.cfgMedCount, ts.cfgMedMinus, ts.cfgMedPlus},
		{"Engineering:", 491, ts.cfgEngineeringCapacity, ts.cfgEngMinus, ts.cfgEngPlus},
		{"Security:", 516, ts.cfgSecurityCapacity, ts.cfgSecMinus, ts.cfgSecPlus},
		{"Life Pod Bays:", 541, ts.cfgLifePodBayCount, ts.cfgPodMinus, ts.cfgPodPlus},
	}
	for _, r := range rows {
		mlge_text.Draw(screen, r.label, 13, cx-150, r.y+5, labelColor)
		r.minus.Draw(screen)
		mlge_text.Draw(screen, fmt.Sprintf("%d", r.val), 13, cx+88, r.y+5, color.RGBA{220, 230, 220, 255})
		r.plus.Draw(screen)
	}

	ts.cfgSelfDestructToggle.Draw(screen)

	ts.confirmNameBtn.Draw(screen)
	ts.cancelNameBtn.Draw(screen)
}

func (ts *TitleScreenState) drawStationBrowser(screen *ebiten.Image) {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2

	heading := "Choose a Station"
	mlge_text.Draw(screen, heading, 20, cx-len(heading)*20*3/10/2, 250, color.RGBA{180, 220, 180, 255})

	if len(ts.stations) == 0 {
		mlge_text.Draw(screen, "No stations found. Generate one first.", 14,
			cx-200, 400, color.RGBA{150, 150, 150, 255})
	} else {
		ts.stationList.Draw(screen)
	}
	ts.selectStationBtn.Draw(screen)
	ts.backFromStationsBtn.Draw(screen)
}

func (ts *TitleScreenState) drawPlayerBrowser(screen *ebiten.Image) {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2

	heading := "Players on: " + ts.selectedStation.Name
	mlge_text.Draw(screen, heading, 18, cx-len(heading)*18*3/10/2, 250, color.RGBA{180, 220, 180, 255})

	if len(ts.players) == 0 {
		mlge_text.Draw(screen, "No player runs yet. Start a new player.", 14,
			cx-200, 400, color.RGBA{150, 150, 150, 255})
	} else {
		ts.playerList.Draw(screen)
	}
	ts.newPlayerBtn.Draw(screen)
	ts.continueBtn.Draw(screen)
	ts.backFromPlayersBtn.Draw(screen)
}

// Done implements client.ClientState.
func (ts *TitleScreenState) Done() bool { return ts.done }
