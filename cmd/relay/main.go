package main

import (
	"log"

	"github.com/john6604/yata-collaborative-editor/internal/relay"
)

const defaultDatabasePath = "../../data/yata.db"

func main() {

	relayServer, errRelay := relay.NewRelayServer(defaultDatabasePath, ":8181")

	if errRelay != nil {
		log.Fatal(errRelay)
	}

	err := relayServer.Start()

	if err != nil {
		log.Fatal(err)
	}
}
