package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlasciiclient"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermclient"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/game"
)

func main() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			os.Chdir(filepath.Dir(exe))
		}
	}

	if err := game.LoadDataHeadless(); err != nil {
		log.Fatal(err)
	}

	sim, err := game.NewSimWorld()
	if err != nil {
		log.Fatal(err)
	}

	srvT, cliT := transport.NewLocalTransport()

	codec := game.NewSPCodec(sim)
	server := simulation.NewServer(
		simulation.ServerConfig{TickRate: 20},
		sim,
		func() []*ecs.Entity { return sim.Level.Entities },
		srvT,
		codec,
	)
	server.SetState(game.NewMainSimState(sim))
	go server.Run()

	asciiWorld := rlasciiclient.NewAsciiWorld()
	tc, err := rltermclient.New(cliT, codec)
	if err != nil {
		server.Stop()
		log.Fatal(err)
	}
	defer tc.Fini()

	// Keep the AsciiWorld in sync with the terminal client's World.
	tc.World = asciiWorld

	// GUI: static HUD + inventory popup + look mode.
	inv := game.NewTermInventoryView(sim)
	look := game.NewTermLookView(sim)
	tc.GUI = &rltermgui.GUI{}
	tc.GUI.Add(game.NewTermHUD(sim))
	tc.GUI.Add(inv)
	tc.GUI.Add(look)

	// Follow the player: center camera on player position each tick.
	tc.OnTick = func(snap *transport.Snapshot) {
		if sim.Player != nil {
			pc := sim.Player.GetComponent("Position").(*component.PositionComponent)
			tc.CameraZ = pc.GetZ()
			cols, rows := tc.ScreenSize()
			tc.CameraX = pc.GetX() - cols/2
			tc.CameraY = pc.GetY() - rows/2
		}
	}

	// Map key events to server commands.
	tc.OnInput = func(ev *tcell.EventKey) *transport.Command {
		// Toggle inventory (GUI is hidden, so this event was not consumed).
		if ev.Rune() == 'i' || ev.Rune() == 'I' {
			inv.Toggle()
			return nil
		}
		key := termKeyToCommand(ev)
		if key == "" {
			return nil
		}
		if key == "quit" {
			return &transport.Command{Type: rltermclient.QuitCommand}
		}
		return &transport.Command{
			Type:    game.CmdAction,
			Payload: game.ActionPayload{Key: key},
		}
	}

	tc.Run()
	server.Stop()
}

// termKeyToCommand maps a tcell key event to a spaceplant command string.
func termKeyToCommand(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyEscape:
		return "quit"
	case tcell.KeyUp:
		return "W"
	case tcell.KeyDown:
		return "S"
	case tcell.KeyLeft:
		return "A"
	case tcell.KeyRight:
		return "D"
	}
	switch ev.Rune() {
	case 'w', 'W':
		return "W"
	case 's', 'S':
		return "S"
	case 'a', 'A':
		return "A"
	case 'd', 'D':
		return "D"
	case '.':
		return "Period"
	case 'p', 'P':
		return "P"
	case 'e', 'E':
		return "E"
	case 'h', 'H':
		return "H"
	case 'q':
		return "quit"
	}
	return ""
}
