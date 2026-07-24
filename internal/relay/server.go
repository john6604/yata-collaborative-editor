package relay

import (
	"net/http"
	"time"
)

type RelayServer struct {
	hub        *Hub
	address    string
	serverHTTP *http.Server
}

func NewRelayServer(address string) *RelayServer {
	hubRelay := NewHub()
	relay := RelayServer{hub: hubRelay}
	relay.address = address
	return &relay
}

func (rs *RelayServer) Start() error {

	mux := http.NewServeMux()

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
