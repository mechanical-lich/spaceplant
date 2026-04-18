package game

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/mechanical-lich/mlge/client"
	mlge_text "github.com/mechanical-lich/mlge/text"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/mlge/ui/minui"
	"github.com/mechanical-lich/spaceplant/internal/config"
)

// TitleScreenState is the initial client state shown before starting or loading a game.
// It implements client.ClientState.
type TitleScreenState struct {
	sim       *SimWorld
	simState  *MainSimState
	transport transport.ClientTransport

	newGameBtn  *minui.Button
	loadGameBtn *minui.Button
	quitBtn     *minui.Button

	saveExists bool
	errMsg     string

	done bool
	next client.ClientState
}

var _ client.ClientState = (*TitleScreenState)(nil)

// NewTitleScreenState creates the title screen.
func NewTitleScreenState(sim *SimWorld, simState *MainSimState, t transport.ClientTransport) *TitleScreenState {
	cfg := config.Global()
	cx := cfg.ScreenWidth / 2

	btnW := 200
	btnH := 36
	btnX := cx - btnW/2

	_, saveErr := os.Stat("save.json")
	saveExists := saveErr == nil

	ts := &TitleScreenState{
		sim:        sim,
		simState:   simState,
		transport:  t,
		saveExists: saveExists,
	}

	// New Game button
	ts.newGameBtn = minui.NewButton("title_new", "New Game")
	ts.newGameBtn.SetPosition(btnX, 340)
	ts.newGameBtn.SetSize(btnW, btnH)
	ts.newGameBtn.OnClick = func() { ts.startNewGame() }

	// Load Game button
	ts.loadGameBtn = minui.NewButton("title_load", "Load Game")
	ts.loadGameBtn.SetPosition(btnX, 340+btnH+12)
	ts.loadGameBtn.SetSize(btnW, btnH)
	ts.loadGameBtn.SetEnabled(saveExists)
	ts.loadGameBtn.OnClick = func() { ts.startLoadGame() }

	// Quit button
	ts.quitBtn = minui.NewButton("title_quit", "Quit")
	ts.quitBtn.SetPosition(btnX, 340+2*(btnH+12))
	ts.quitBtn.SetSize(btnW, btnH)
	ts.quitBtn.OnClick = func() { os.Exit(0) }

	return ts
}

func (ts *TitleScreenState) startNewGame() {
	ts.sim.RegenerateLevel()
	ts.simState.ResetPhase()
	ts.next = NewSPClientState(ts.sim, ts.simState, ts.transport)
	ts.done = true
}

func (ts *TitleScreenState) startLoadGame() {
	if err := LoadIntoSimWorld(ts.sim, "save.json"); err != nil {
		ts.errMsg = "Load failed: " + err.Error()
		return
	}
	ts.simState.ResetPhase()
	ts.next = NewSPClientState(ts.sim, ts.simState, ts.transport)
	ts.done = true
}

// Update implements client.ClientState.
func (ts *TitleScreenState) Update(_ *transport.Snapshot) client.ClientState {
	// Keyboard shortcuts: N = New Game, L = Load Game, Q/Escape = Quit.
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		ts.startNewGame()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) && ts.saveExists {
		ts.startLoadGame()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	ts.newGameBtn.Update()
	ts.loadGameBtn.Update()
	ts.quitBtn.Update()

	return ts.next
}

// Draw implements client.ClientState.
func (ts *TitleScreenState) Draw(screen *ebiten.Image) {
	cfg := config.Global()
	w := float64(cfg.ScreenWidth)
	h := float64(cfg.ScreenHeight)

	// Dark background.
	ebitenutil.DrawRect(screen, 0, 0, w, h, color.RGBA{10, 10, 18, 255})

	// Title.
	titleText := "Space Plants!"
	mlge_text.Draw(screen, titleText, 52, cfg.ScreenWidth/2-len(titleText)*52*3/10/2, 180, color.RGBA{180, 230, 120, 255})

	// Subtitle.
	subText := "A Sci-Fi Roguelike"
	mlge_text.Draw(screen, subText, 18, cfg.ScreenWidth/2-len(subText)*18*3/10/2, 250, color.RGBA{130, 160, 100, 255})

	// Buttons.
	ts.newGameBtn.Draw(screen)
	ts.loadGameBtn.Draw(screen)
	ts.quitBtn.Draw(screen)

	// Error message (e.g. load failed).
	if ts.errMsg != "" {
		mlge_text.Draw(screen, ts.errMsg, 13, cfg.ScreenWidth/2-200, 560, color.RGBA{220, 80, 80, 255})
	}

	// Key hints at bottom.
	hint := "[N] New Game    [L] Load Game    [Q] Quit"
	mlge_text.Draw(screen, hint, 12, cfg.ScreenWidth/2-len(hint)*12*3/10/2, int(h)-40, color.RGBA{90, 100, 90, 255})
}

// Done implements client.ClientState.
func (ts *TitleScreenState) Done() bool { return ts.done }
