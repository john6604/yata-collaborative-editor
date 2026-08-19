package relay

import (
	"net/http"
	"time"

	"github.com/john6604/yata-collaborative-editor/internal/storage"
)

type RelayServer struct {
	Hub        *Hub
	Address    string
	ServerHTTP *http.Server
	Storage    *storage.Storage
}

func NewRelayServer(path string, address string) (*RelayServer, error) {
	var newStorage storage.Storage
	errStorage := newStorage.OpenDB(path)
	if errStorage != nil {
		return &RelayServer{}, errStorage
	}
	hubRelay := NewHub()
	relay := RelayServer{Hub: hubRelay}
	relay.Address = address
	relay.Storage = &newStorage
	return &relay, nil
}

func (rs *RelayServer) Start(mux *http.ServeMux) error {

	server := &http.Server{
		Addr:           rs.Address,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	rs.ServerHTTP = server

	return server.ListenAndServe()
}

func (rs *RelayServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", run)
	mux.HandleFunc("/ws", rs.ws)
}
