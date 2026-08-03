package relay

import (
	"net/http"
	"time"

	"github.com/john6604/yata-collaborative-editor/internal/storage"
)

type RelayServer struct {
	hub        *Hub
	address    string
	serverHTTP *http.Server
	storage    *storage.Storage
}

func NewRelayServer(path string, address string) (*RelayServer, error) {
	var newStorage storage.Storage
	errStorage := newStorage.OpenDB(path)
	if errStorage != nil {
		return &RelayServer{}, errStorage
	}
	hubRelay := NewHub()
	relay := RelayServer{hub: hubRelay}
	relay.address = address
	relay.storage = &newStorage
	return &relay, nil
}

func (rs *RelayServer) Start() error {

	mux := http.NewServeMux()
	defer rs.storage.CloseDB()

	mux.HandleFunc("/", run)
	mux.HandleFunc("/ws", rs.ws)

	server := &http.Server{
		Addr:           rs.address,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	rs.serverHTTP = server

	return server.ListenAndServe()
}
