package main

import (
	"log"

	"github.com/mechanical-lich/spaceplant/game"
)

func main() {
	g, err := game.NewGame("Space Plant!")
	if err != nil {
		log.Fatal(err)
	}

	err = g.Run()

	if err != nil {
		log.Fatal(err)
	}
}
