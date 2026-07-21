package relay

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

type RelayServer struct {
	hub        *Hub
	address    string
	serverHTTP *http.Server
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewRelayServer(address string) *RelayServer {
	hubRelay := NewHub()
	relay := RelayServer{hub: hubRelay}
	relay.address = address
	return &relay
}

func run(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(response, "Relay Running...")
}

func (rs *RelayServer) ws(response http.ResponseWriter, request *http.Request) {

	conn, err := upgrader.Upgrade(response, request, nil)

	if err != nil {
		return
	}

	defer conn.Close()

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		return
	}

	if messageType != websocket.TextMessage {
		return
	}

	version, requestType, payloadJoin, err := protocol.DecodeJoinMessage(message)

	if err != nil {
		return
	}

	if version != protocol.SupportedVersion {
		return
	}

	if requestType != protocol.TypeJoin {
		return
	}

	room, client, errJoin := protocol.DecodeJoin(payloadJoin)

	if errJoin != nil {
		return
	}

	session, err1 := rs.hub.Join(room, client, conn)

	if err1 != nil {
		return
	}

	defer rs.hub.Leave(session)

	ackBytes, errAck := protocol.EncodeJoinAck(room, client)

	if errAck != nil {
		return
	}

	errSend := conn.WriteMessage(websocket.TextMessage, ackBytes)

	if errSend != nil {
		return
	}

	for {

		messageType, _, err := conn.ReadMessage()

		if err != nil {
			break
		}

		if messageType != websocket.TextMessage {
			break
		}

	}
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
