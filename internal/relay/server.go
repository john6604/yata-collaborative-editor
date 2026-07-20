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

func handler(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Relay running..."))
}

func (rs *RelayServer) Start() error {

	myHandler := http.HandlerFunc(handler)

	server := &http.Server{
		Addr:           rs.address,
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	rs.serverHTTP = server

	return server.ListenAndServe()
}
