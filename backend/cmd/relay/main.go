package main

import (
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/john6604/yata-collaborative-editor/internal/catalog"
	"github.com/john6604/yata-collaborative-editor/internal/httpapi"
	"github.com/john6604/yata-collaborative-editor/internal/relay"
)

const defaultDatabasePathYata = "../../data/yata.db"

func defaultDatabasePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("data", "catalog.db")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "data", "catalog.db"))
}

func main() {

	relayServer, errRelay := relay.NewRelayServer(defaultDatabasePathYata, ":8181")

	if errRelay != nil {
		log.Fatal(errRelay)
	}

	store, errStore := catalog.NewStore(defaultDatabasePath())
	if errStore != nil {
		log.Fatal(errStore)
	}

	api := httpapi.NewAPI(store)

	mux := http.NewServeMux()

	api.RegisterRoutes(mux)
	relayServer.RegisterRoutes(mux)

	err := relayServer.Start(mux)

	if err != nil {
		relayServer.Storage.CloseDB()
		log.Fatal(err)
	}
}
