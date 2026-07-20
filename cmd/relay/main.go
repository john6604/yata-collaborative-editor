package main

import (
	"log"

	"github.com/john6604/yata-collaborative-editor/internal/relay"
)

func main() {

	relayServer := relay.NewRelayServer(":8181")

	err := relayServer.Start()

	if err != nil {
		log.Fatal(err)
	}
}
