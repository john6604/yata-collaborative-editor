package relay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
	"github.com/john6604/yata-collaborative-editor/internal/storage"
	internalSync "github.com/john6604/yata-collaborative-editor/internal/sync"
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
		SendErrorMessage(
			protocol.ExpectedMessageCode,
			protocol.ExpectedMessage,
			conn,
		)
		return
	}

	version, requestType, payloadJoin, err := protocol.DecodeEnvelope(message)
	if err != nil {
		SendErrorMessage(
			protocol.InvalidPayload,
			protocol.InvalidPayloadMessage,
			conn,
		)
		return
	}

	if version != protocol.SupportedVersion {
		SendErrorMessage(
			protocol.UnsupportedVersion,
			protocol.UnsupportedVersionMessage,
			conn,
		)
		return
	}

	if requestType != protocol.TypeJoin {
		SendErrorMessage(
			protocol.ExpectedJoin,
			protocol.ExpectedJoinMessage,
			conn,
		)
		return
	}

	room, client, errJoin := protocol.DecodeJoin(payloadJoin)
	if errJoin != nil {
		SendErrorMessage(
			protocol.InvalidPayload,
			protocol.InvalidPayloadMessage,
			conn,
		)
		return
	}

	session, errJoinHub := rs.hub.Join(room, client, conn)
	if errJoinHub != nil {
		SendErrorMessage(
			protocol.InternalError,
			protocol.InternalErrorMessage,
			conn,
		)
		return
	}

	defer rs.hub.Leave(session)

	ackBytes, errAck := protocol.EncodeJoinAck(room, client)
	if errAck != nil {
		SendErrorMessage(
			protocol.InternalError,
			protocol.InternalErrorMessage,
			conn,
		)
		return
	}

	errSend := conn.WriteMessage(websocket.TextMessage, ackBytes)
	if errSend != nil {
		return
	}

	for {
		messageType, message, err = conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType != websocket.TextMessage {
			SendErrorMessage(
				protocol.ExpectedMessageCode,
				protocol.ExpectedMessage,
				conn,
			)
			break
		}

		version, typeMessage, payload, errEnvelope := protocol.DecodeEnvelope(message)
		if errEnvelope != nil {
			SendErrorMessage(
				protocol.InvalidPayload,
				protocol.InvalidPayloadMessage,
				conn,
			)
			continue
		}

		if version != protocol.SupportedVersion {
			SendErrorMessage(
				protocol.UnsupportedVersion,
				protocol.UnsupportedVersionMessage,
				conn,
			)
			continue
		}

		switch typeMessage {
		case protocol.TypeUpdate:
			update, errPayload := protocol.DecodeUpdate(payload)
			if errPayload != nil {
				SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
					conn,
				)
				continue
			}

			var operationEnvelope protocol.OperationEnvelope

			errDecode := json.Unmarshal(
				update.Operation,
				&operationEnvelope,
			)
			if errDecode != nil {
				SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
					conn,
				)
				continue
			}

			formattedType := strings.TrimSpace(operationEnvelope.Type)
			if formattedType == "" {
				SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
					conn,
				)
				continue
			}

			if formattedType == protocol.OpSnapshot {
				snapshot, errSnapshot := protocol.DecodeSnapshot(update)
				if errSnapshot != nil {
					SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage, conn)
					continue
				}
				savingRoom := session.roomID
				errDelta := rs.storage.SaveSnapshotRoom(savingRoom, *snapshot.Delta)
				if errDelta != nil {
					SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
					continue
				}
				continue
			}

			receiversExist, errBroadcast := rs.hub.BroadcastToRoom(
				session,
				message,
			)
			if errBroadcast != nil {
				if errBroadcast == ErrNotJoined {
					SendErrorMessage(
						protocol.NotJoined,
						protocol.NotJoinedMessage,
						conn,
					)
				} else {
					SendErrorMessage(
						protocol.InternalError,
						protocol.InternalErrorMessage,
						conn,
					)
				}

				continue
			}

			if !receiversExist && formattedType == protocol.OpSync1 {
				var emptyDelta protocol.Delta

				delta, errDelta := rs.storage.LoadInternalSnapshot(session.roomID)
				if errDelta != nil {
					if errors.Is(errDelta, storage.ErrSnapshotNotFound) {
						sync2, errSync2 := internalSync.EncodeSyncStep2(emptyDelta)
						if errSync2 != nil {
							SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
							continue
						}

						if errSendSync2 := session.Send(sync2); errSendSync2 != nil {
							return
						}
						continue
					}
					SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
					continue
				}

				sync2, errSync2 := internalSync.EncodeSyncStep2(delta)
				if errSync2 != nil {
					SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage, conn)
					continue
				}

				if errSendSync2 := session.Send(sync2); errSendSync2 != nil {
					return
				}
			}

			continue

		default:
			SendErrorMessage(protocol.UnknownMessageCode, protocol.UnknownMessage, conn)
			continue
		}
	}
}
