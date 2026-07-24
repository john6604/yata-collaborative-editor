package relay

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

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
		SendErrorMessage(protocol.ExpectedMessageCode, protocol.ExpectedMessage, conn)
		return
	}

	version, requestType, payloadJoin, err := protocol.DecodeEnvelope(message)

	if err != nil {
		SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage, conn)
		return
	}

	if version != protocol.SupportedVersion {
		SendErrorMessage(protocol.UnsupportedVersion, protocol.UnsupportedVersionMessage, conn)
		return
	}

	if requestType != protocol.TypeJoin {
		SendErrorMessage(protocol.ExpectedJoin, protocol.ExpectedJoinMessage, conn)
		return
	}

	room, client, errJoin := protocol.DecodeJoin(payloadJoin)

	if errJoin != nil {
		SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage, conn)
		return
	}

	session, err1 := rs.hub.Join(room, client, conn)

	if err1 != nil {
		SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
		return
	}

	defer rs.hub.Leave(session)

	ackBytes, errAck := protocol.EncodeJoinAck(room, client)

	if errAck != nil {
		SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
		return
	}

	errSend := conn.WriteMessage(websocket.TextMessage, ackBytes)

	if errSend != nil {
		return
	}

	for {

		messageType, message, err := conn.ReadMessage()

		if err != nil {
			break
		}

		if messageType != websocket.TextMessage {
			SendErrorMessage(protocol.ExpectedMessageCode, protocol.ExpectedMessage, conn)
			break
		}

		version, typeMessage, bytes, errBytes := protocol.DecodeEnvelope(message)

		if errBytes != nil {
			SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage, conn)
			continue
		}

		if version != protocol.SupportedVersion {
			SendErrorMessage(protocol.UnsupportedVersion, protocol.UnsupportedVersionMessage, conn)
			continue
		}

		switch typeMessage {
		case protocol.TypeUpdate:

			_, errPayload := protocol.DecodeUpdate(bytes)

			if errPayload != nil {
				SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage, conn)
				continue
			}

			err := rs.hub.BroadcastToRoom(session, message)

			if err != nil {
				if err == ErrNotJoined {
					SendErrorMessage(protocol.NotJoined, protocol.NotJoinedMessage, conn)
				} else {
					SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
				}
				continue
			}
			continue
		default:
			SendErrorMessage(protocol.UnknownMessageCode, protocol.UnknownMessage, conn)
			continue
		}
	}
}
