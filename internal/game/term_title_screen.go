package game

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/spaceplant/internal/lore"
)

type termTitleScreen int

const (
	termScreenMain          termTitleScreen = iota
	termScreenNewStationName                // enter / randomise station name
	termScreenStationBrowser
	termScreenPlayerBrowser
)

// TermTitleScreen is the terminal title screen.
// It blocks all input until the player has chosen to start / load a run.
// Callbacks are set by cmd/terminal/main.go.
type TermTitleScreen struct {
	visible bool
	Quit    bool // true if user dismissed via Esc without selecting anything
	screen  termTitleScreen

	// Populated once the user has made a selection.
	OnNewStation     func(name string)           // generate new station with given name
	OnLoadStation    func(stationID string)       // load station, then show character creator
	OnContinuePlayer func(stationID, playerRunID string) // load full game

	// Station browser
	stations       []StationMeta
	stationCursor  int

	// Player browser
	players       []PlayerRunMeta
	playerCursor  int
	selectedStation StationMeta

	// Name input for new station
	nameInput   []rune
	errMsg      string
}

func NewTermTitleScreen() *TermTitleScreen {
	return &TermTitleScreen{visible: true, screen: termScreenMain}
}

func (v *TermTitleScreen) Visible() bool { return v.visible }
func (v *TermTitleScreen) Hide()         { v.visible = false }

// HandleKey implements rltermgui.View. Consumes all input while visible.
func (v *TermTitleScreen) HandleKey(ev *tcell.EventKey) bool {
	if !v.visible {
		return false
	}
	switch v.screen {
	case termScreenMain:
		v.handleMain(ev)
	case termScreenNewStationName:
		v.handleNameInput(ev)
	case termScreenStationBrowser:
		v.handleStationBrowser(ev)
	case termScreenPlayerBrowser:
		v.handlePlayerBrowser(ev)
	}
	return true
}

func (v *TermTitleScreen) handleMain(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'n', 'N':
			v.nameInput = nil
			v.errMsg = ""
			v.screen = termScreenNewStationName
		case 'b', 'B':
			v.openStationBrowser()
		case 'q', 'Q':
			v.screen = termScreenMain // will quit via Escape below
		}
	case tcell.KeyEscape:
		v.Quit = true
		v.visible = false
	}
}

func (v *TermTitleScreen) handleNameInput(ev *tcell.EventKey) {
	v.errMsg = ""
	switch ev.Key() {
	case tcell.KeyEscape:
		v.screen = termScreenMain
	case tcell.KeyEnter:
		name := string(v.nameInput)
		if name == "" {
			name = lore.RandomStationName()
		}
		if v.OnNewStation != nil {
			v.OnNewStation(name)
		}
		v.visible = false
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(v.nameInput) > 0 {
			v.nameInput = v.nameInput[:len(v.nameInput)-1]
		}
	case tcell.KeyRune:
		r := ev.Rune()
		switch r {
		case 0x12: // Ctrl+R — randomise
			v.nameInput = []rune(lore.RandomStationName())
		default:
			if len(v.nameInput) < 40 {
				v.nameInput = append(v.nameInput, r)
			}
		}
	}
}

func (v *TermTitleScreen) openStationBrowser() {
	stations, err := ListStations(savesDir)
	if err != nil {
		v.errMsg = "Error loading stations: " + err.Error()
		return
	}
	v.stations = stations
	v.stationCursor = 0
	v.errMsg = ""
	v.screen = termScreenStationBrowser
}

func (v *TermTitleScreen) handleStationBrowser(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		v.screen = termScreenMain
	case tcell.KeyUp:
		if v.stationCursor > 0 {
			v.stationCursor--
		}
	case tcell.KeyDown:
		if v.stationCursor < len(v.stations)-1 {
			v.stationCursor++
		}
	case tcell.KeyEnter:
		if len(v.stations) == 0 {
			return
		}
		v.selectedStation = v.stations[v.stationCursor]
		v.openPlayerBrowser()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'n', 'N':
			// new player on selected station
			if len(v.stations) == 0 {
				return
			}
			v.selectedStation = v.stations[v.stationCursor]
			if v.OnLoadStation != nil {
				v.OnLoadStation(v.selectedStation.StationID)
			}
			v.visible = false
		}
	}
}

func (v *TermTitleScreen) openPlayerBrowser() {
	players, err := ListPlayerRuns(savesDir, v.selectedStation.StationID)
	if err != nil {
		v.errMsg = "Error loading players: " + err.Error()
		return
	}
	v.players = players
	v.playerCursor = 0
	v.errMsg = ""
	v.screen = termScreenPlayerBrowser
}

func (v *TermTitleScreen) handlePlayerBrowser(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		v.openStationBrowser()
	case tcell.KeyUp:
		if v.playerCursor > 0 {
			v.playerCursor--
		}
	case tcell.KeyDown:
		if v.playerCursor < len(v.players)-1 {
			v.playerCursor++
		}
	case tcell.KeyEnter:
		v.continueSelected()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'n', 'N':
			if v.OnLoadStation != nil {
				v.OnLoadStation(v.selectedStation.StationID)
			}
			v.visible = false
		case 'c', 'C':
			v.continueSelected()
		}
	}
}

func (v *TermTitleScreen) continueSelected() {
	if len(v.players) == 0 {
		return
	}
	p := v.players[v.playerCursor]
	if p.Dead {
		v.errMsg = "That player is dead — start a new one with [N]."
		return
	}
	if v.OnContinuePlayer != nil {
		v.OnContinuePlayer(p.StationID, p.PlayerRunID)
	}
	v.visible = false
}

// Draw implements rltermgui.View.
func (v *TermTitleScreen) Draw(s tcell.Screen) {
	if !v.visible {
		return
	}
	sw, sh := s.Size()

	// Fill background.
	bg := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	rltermgui.FillRect(s, 0, 0, sw, sh, bg)

	// Title banner.
	title := "S P A C E   P L A N T S !"
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack).Bold(true)
	rltermgui.DrawText(s, (sw-len(title))/2, sh/2-10, title, titleStyle)
	sub := "A Sci-Fi Roguelike"
	subStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGreen).Background(tcell.ColorBlack)
	rltermgui.DrawText(s, (sw-len(sub))/2, sh/2-8, sub, subStyle)

	switch v.screen {
	case termScreenMain:
		v.drawMain(s, sw, sh)
	case termScreenNewStationName:
		v.drawNameInput(s, sw, sh)
	case termScreenStationBrowser:
		v.drawStationBrowser(s, sw, sh)
	case termScreenPlayerBrowser:
		v.drawPlayerBrowser(s, sw, sh)
	}

	if v.errMsg != "" {
		errStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
		rltermgui.DrawText(s, (sw-len(v.errMsg))/2, sh-3, v.errMsg, errStyle)
	}
}

func (v *TermTitleScreen) drawMain(s tcell.Screen, sw, sh int) {
	items := []string{
		"[N] New Station",
		"[B] Browse Stations",
		"[Q/Esc] Quit",
	}
	y := sh/2 - 3
	menuStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	for _, item := range items {
		rltermgui.DrawText(s, (sw-len(item))/2, y, item, menuStyle)
		y++
	}
}

func (v *TermTitleScreen) drawNameInput(s tcell.Screen, sw, sh int) {
	border := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	content := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	hint := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)
	cursor := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	w := 50
	h := 7
	x := (sw - w) / 2
	y := sh/2 - h/2
	rltermgui.FillRect(s, x, y, w, h, content)
	rltermgui.DrawBox(s, x, y, w, h, " New Station Name ", border)

	label := "Name: "
	rltermgui.DrawText(s, x+2, y+2, label, content)
	nameStr := string(v.nameInput)
	rltermgui.DrawText(s, x+2+len(label), y+2, nameStr, content)
	// Draw cursor block at end of text.
	rltermgui.DrawText(s, x+2+len(label)+len(nameStr), y+2, " ", cursor)

	rltermgui.DrawText(s, x+2, y+h-2, "[Enter] confirm  [Ctrl+R] random  [Esc] back", hint)
}

func (v *TermTitleScreen) drawStationBrowser(s tcell.Screen, sw, sh int) {
	border := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	content := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	sel := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	hint := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)

	w := sw * 2 / 3
	if w < 50 {
		w = 50
	}
	maxRows := sh - 12
	h := len(v.stations) + 4
	if h < 6 {
		h = 6
	}
	if h > maxRows {
		h = maxRows
	}
	x := (sw - w) / 2
	y := sh/2 - h/2

	rltermgui.FillRect(s, x, y, w, h, content)
	rltermgui.DrawBox(s, x, y, w, h, " Choose a Station ", border)

	if len(v.stations) == 0 {
		rltermgui.DrawText(s, x+2, y+2, "No stations found — press [Esc] and use [N] to generate one.", dim)
	} else {
		maxVisible := h - 4
		start := 0
		if v.stationCursor >= maxVisible {
			start = v.stationCursor - maxVisible + 1
		}
		for i := 0; i < maxVisible && start+i < len(v.stations); i++ {
			idx := start + i
			st := v.stations[idx]
			label := fmt.Sprintf("  %s", st.Name)
			style := content
			if idx == v.stationCursor {
				label = fmt.Sprintf("> %s", st.Name)
				style = sel
			}
			if len([]rune(label)) > w-2 {
				label = string([]rune(label)[:w-2])
			}
			rltermgui.DrawText(s, x+1, y+2+i, label, style)
		}
	}
	rltermgui.DrawText(s, x+2, y+h-2, "[↑↓] select  [Enter] view players  [N] new player  [Esc] back", hint)
}

func (v *TermTitleScreen) drawPlayerBrowser(s tcell.Screen, sw, sh int) {
	border := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	content := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	sel := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	dead := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	hint := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)

	w := sw * 2 / 3
	if w < 50 {
		w = 50
	}
	maxRows := sh - 12
	h := len(v.players) + 5
	if h < 7 {
		h = 7
	}
	if h > maxRows {
		h = maxRows
	}
	x := (sw - w) / 2
	y := sh/2 - h/2

	title := fmt.Sprintf(" Players on: %s ", v.selectedStation.Name)
	rltermgui.FillRect(s, x, y, w, h, content)
	rltermgui.DrawBox(s, x, y, w, h, title, border)

	if len(v.players) == 0 {
		dimStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
		rltermgui.DrawText(s, x+2, y+2, "No player runs yet — press [N] to start one.", dimStyle)
	} else {
		maxVisible := h - 5
		start := 0
		if v.playerCursor >= maxVisible {
			start = v.playerCursor - maxVisible + 1
		}
		for i := 0; i < maxVisible && start+i < len(v.players); i++ {
			idx := start + i
			p := v.players[idx]
			label := p.Name
			if p.ClassName != "" {
				label += " — " + p.ClassName
			}
			label += fmt.Sprintf(" — Floor %d", p.CurrentZ+1)
			if p.Dead {
				label += " (dead)"
			}
			prefix := "  "
			style := content
			if p.Dead {
				style = dead
			}
			if idx == v.playerCursor {
				prefix = "> "
				if !p.Dead {
					style = sel
				}
			}
			line := prefix + label
			if len([]rune(line)) > w-2 {
				line = string([]rune(line)[:w-2])
			}
			rltermgui.DrawText(s, x+1, y+2+i, line, style)
		}
	}
	rltermgui.DrawText(s, x+2, y+h-2, "[↑↓] select  [C/Enter] continue  [N] new player  [Esc] back", hint)
}
