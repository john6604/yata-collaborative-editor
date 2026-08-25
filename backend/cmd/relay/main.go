package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/john6604/yata-collaborative-editor/internal/catalog"
	"github.com/john6604/yata-collaborative-editor/internal/httpapi"
	"github.com/john6604/yata-collaborative-editor/internal/relay"
)

const dataDirEnvVar = "DATA_DIR"

type databasePaths struct {
	dataDir string
	yata    string
	catalog string
}

func defaultDataDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(".", "data")
	}

	return filepath.Join(filepath.Dir(file), "..", "..", "data")
}

func resolveDataDir() (string, error) {
	dataDir := strings.TrimSpace(os.Getenv(dataDirEnvVar))
	if dataDir == "" {
		dataDir = defaultDataDir()
	}

	absoluteDataDir, err := filepath.Abs(filepath.Clean(dataDir))
	if err != nil {
		return "", err
	}

	return absoluteDataDir, nil
}

func resolveDatabasePaths() (databasePaths, error) {
	dataDir, err := resolveDataDir()
	if err != nil {
		return databasePaths{}, err
	}

	if errMkdir := os.MkdirAll(dataDir, 0o755); errMkdir != nil {
		return databasePaths{}, fmt.Errorf("create data directory %q: %w", dataDir, errMkdir)
	}

	return databasePaths{
		dataDir: dataDir,
		yata:    filepath.Join(dataDir, "yata.db"),
		catalog: filepath.Join(dataDir, "catalog.db"),
	}, nil
}

func main() {

	paths, errPaths := resolveDatabasePaths()
	if errPaths != nil {
		log.Fatal(errPaths)
	}

	relayServer, errRelay := relay.NewRelayServer(paths.yata, ":8181")

	if errRelay != nil {
		log.Fatal(errRelay)
	}

	store, errStore := catalog.NewStore(paths.catalog)
	if errStore != nil {
		relayServer.Storage.CloseDB()
		log.Fatal(errStore)
	}

	api := httpapi.NewAPI(store)

	mux := http.NewServeMux()

	api.RegisterRoutes(mux)
	relayServer.RegisterRoutes(mux)

	err := relayServer.Start(mux)

	if err != nil {
		relayServer.Storage.CloseDB()
		store.Close()
		log.Fatal(err)
	}
}
