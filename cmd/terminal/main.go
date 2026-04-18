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
	"github.com/mechanical-lich/mlge/message"
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

	asciiWorld := rlasciiclient.NewAsciiWorld()
	tc, err := rltermclient.New(cliT, codec)
	if err != nil {
		server.Stop()
		log.Fatal(err)
	}
	defer tc.Fini()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				tc.Fini()
				panic(r)
			}
		}()
		server.Run()
	}()

	tc.World = asciiWorld

	// Modal views — created once; modals are nil until player spawns.
	termCC := game.NewTermCharacterCreator()
	hud := game.NewTermHUD(sim)
	inv := game.NewTermInventoryView(sim)
	look := game.NewTermLookView(sim)

	var reload *game.TermReloadView
	var aimedShot *game.TermAimedShotView
	var loot *game.TermLootView
	var classView *game.TermClassView

	termCC.OnComplete = func(data game.CharacterData) {
		if err := sim.SpawnPlayer(data); err != nil {
			message.AddMessage("Error spawning player: " + err.Error())
			return
		}
		reload = game.NewTermReloadView(sim.Player)
		reload.OnReload = func(weaponItem, ammoItem *ecs.Entity) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdReload,
				Payload: game.ReloadPayload{WeaponItem: weaponItem, AmmoItem: ammoItem},
			})
		}
		tc.GUI.Add(reload)

		aimedShot = game.NewTermAimedShotView()
		aimedShot.OnSelect = func(bodyPart string) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdAimedShot,
				Payload: game.AimedShotPayload{BodyPart: bodyPart},
			})
		}
		tc.GUI.Add(aimedShot)

		loot = game.NewTermLootView()
		loot.OnPickup = func(item *ecs.Entity, tx, ty, tz int) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdPickupItem,
				Payload: game.PickupItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
			})
		}
		loot.OnEquip = func(item *ecs.Entity, tx, ty, tz int) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdEquipItem,
				Payload: game.EquipItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
			})
		}
		tc.GUI.Add(loot)

		classView = game.NewTermClassView(sim.Player)
		tc.GUI.Add(classView)
	}

	titleScreen := game.NewTermTitleScreen()
	titleScreen.OnNewStation = func(name string) {
		if err := sim.RegenerateLevel(); err != nil {
			log.Printf("RegenerateLevel: %v", err)
			return
		}
		sim.StationName = name
		if err := game.SaveStation(sim, "saves"); err != nil {
			log.Printf("SaveStation: %v", err)
		}
		termCC.Activate()
	}
	titleScreen.OnLoadStation = func(stationID string) {
		if err := game.LoadStationIntoSimWorld(sim, stationID, "saves"); err != nil {
			log.Printf("LoadStationIntoSimWorld: %v", err)
		}
		termCC.Activate()
	}
	titleScreen.OnContinuePlayer = func(stationID, playerRunID string) {
		if err := game.LoadFullGame(sim, stationID, playerRunID, "saves"); err != nil {
			log.Printf("LoadFullGame: %v", err)
			return
		}
		// Player already exists; set up game views.
		reload = game.NewTermReloadView(sim.Player)
		reload.OnReload = func(weaponItem, ammoItem *ecs.Entity) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdReload,
				Payload: game.ReloadPayload{WeaponItem: weaponItem, AmmoItem: ammoItem},
			})
		}
		tc.GUI.Add(reload)
		aimedShot = game.NewTermAimedShotView()
		aimedShot.OnSelect = func(bodyPart string) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdAimedShot,
				Payload: game.AimedShotPayload{BodyPart: bodyPart},
			})
		}
		tc.GUI.Add(aimedShot)
		loot = game.NewTermLootView()
		loot.OnPickup = func(item *ecs.Entity, tx, ty, tz int) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdPickupItem,
				Payload: game.PickupItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
			})
		}
		loot.OnEquip = func(item *ecs.Entity, tx, ty, tz int) {
			cliT.SendCommand(&transport.Command{
				Type:    game.CmdEquipItem,
				Payload: game.EquipItemPayload{Item: item, TileX: tx, TileY: ty, TileZ: tz},
			})
		}
		tc.GUI.Add(loot)
		classView = game.NewTermClassView(sim.Player)
		tc.GUI.Add(classView)
	}

	tc.GUI = &rltermgui.GUI{}
	tc.GUI.Add(titleScreen)
	tc.GUI.Add(termCC) // character creator shown after title screen exits
	tc.GUI.Add(hud)
	tc.GUI.Add(inv)
	tc.GUI.Add(look)

	tc.OnTick = func(snap *transport.Snapshot) {
		if titleScreen.Quit {
			cliT.SendCommand(&transport.Command{Type: rltermclient.QuitCommand})
			return
		}
		if termCC.Active() {
			return
		}
		if sim.Player != nil {
			pc := sim.Player.GetComponent("Position").(*component.PositionComponent)
			tc.CameraZ = pc.GetZ()
			cols, rows := tc.ScreenSize()
			tc.CameraX = pc.GetX() - cols/2
			tc.CameraY = pc.GetY() - rows/2
		}
	}

	tc.OnInput = func(ev *tcell.EventKey) *transport.Command {
		// Character creator and other GUI views handled by GUI.HandleKey above OnInput.
		// (tc internals call GUI.HandleKey before OnInput, so this is just game input.)

		// Inventory toggle.
		if ev.Rune() == 'i' || ev.Rune() == 'I' {
			inv.Toggle()
			return nil
		}

		// Class upgrade modal.
		if ev.Rune() == 'C' {
			if classView != nil {
				classView.Open()
			}
			return nil
		}

		// Nearby loot.
		if ev.Rune() == 'p' || ev.Rune() == 'P' {
			if loot != nil && game.TermHasNearbyItems(sim.Player, sim.Level) {
				loot.Open(sim.Player, sim.Level)
				return nil
			}
		}

		// Shift+R → reload modal.
		if ev.Rune() == 'R' {
			if reload != nil {
				reload.Open()
			}
			return nil
		}

		// Shift+F → aimed shot modal.
		if ev.Rune() == 'F' {
			if aimedShot != nil {
				target := game.TermRayTarget(sim.Player, sim.Level)
				if target != nil {
					aimedShot.Open(target)
				} else {
					message.AddMessage("Nothing to aim at.")
				}
			}
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
		return "w"
	case tcell.KeyDown:
		return "s"
	case tcell.KeyLeft:
		return "a"
	case tcell.KeyRight:
		return "d"
	}
	switch ev.Rune() {
	case 'w', 'W':
		return "w"
	case 's', 'S':
		return "s"
	case 'a', 'A':
		return "a"
	case 'd', 'D':
		return "d"
	case '.':
		return "Period"
	case 'p':
		return "p"
	case 'e', 'E':
		return "e"
	case 'h', 'H':
		return "h"
	case 'r':
		return "r" // rush toggle
	case 'g', 'G':
		return "g" // burst fire
	case 'f':
		return "f" // snap shot
	case 'q':
		return "quit"
	}
	return ""
}
