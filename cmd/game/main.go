package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mechanical-lich/mlge/client"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/simulation"
	"github.com/mechanical-lich/mlge/transport"
	"github.com/mechanical-lich/spaceplant/internal/config"
	"github.com/mechanical-lich/spaceplant/internal/game"
)

func main() {
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			os.Chdir(filepath.Dir(exe))
		}
	}

	if err := game.LoadData(); err != nil {
		log.Fatal(err)
	}

	sim, err := game.NewSimWorld()
	if err != nil {
		log.Fatal(err)
	}

	srvT, cliT := transport.NewLocalTransport()

	codec := &noopCodec{}
	server := simulation.NewServer(
		simulation.ServerConfig{TickRate: 20},
		sim,
		func() []*ecs.Entity { return sim.Level.Entities },
		srvT,
		codec,
	)
	server.SetState(game.NewMainSimState(sim))
	go server.Run()

	clientState := game.NewSPClientState(sim, cliT)
	c := client.NewClient(cliT, codec, clientState, sim, func() []*ecs.Entity { return sim.Level.Entities }, client.ClientConfig{
		ScreenWidth:  config.ScreenWidth,
		ScreenHeight: config.ScreenHeight,
		WindowTitle:  "Space Plant!",
	})

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
	server.Stop()
}

// noopCodec satisfies transport.SnapshotCodec.
// The graphical client renders directly from the shared SimWorld.
type noopCodec struct{}

func (n *noopCodec) Encode(tick uint64, _ []*ecs.Entity) *transport.Snapshot {
	return transport.NewSnapshot(tick, nil)
}
func (n *noopCodec) Decode(_ *transport.Snapshot, _ any) {}
