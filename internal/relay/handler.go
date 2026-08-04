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
		session.SendErrorMessage(
			protocol.InternalError,
			protocol.InternalErrorMessage,
		)
		return
	}

	errSend := session.Send(ackBytes)
	if errSend != nil {
		return
	}

	for {
		messageType, message, err = conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType != websocket.TextMessage {
			errSend := session.SendErrorMessage(
				protocol.ExpectedMessageCode,
				protocol.ExpectedMessage,
			)
			if errSend != nil {
				return
			}
			break
		}

		version, typeMessage, payload, errEnvelope := protocol.DecodeEnvelope(message)
		if errEnvelope != nil {
			errSend := session.SendErrorMessage(
				protocol.InvalidPayload,
				protocol.InvalidPayloadMessage,
			)
			if errSend != nil {
				return
			}
			continue
		}

		if version != protocol.SupportedVersion {
			errSend := session.SendErrorMessage(
				protocol.UnsupportedVersion,
				protocol.UnsupportedVersionMessage,
			)
			if errSend != nil {
				return
			}
			continue
		}

		switch typeMessage {
		case protocol.TypeUpdate:
			update, errPayload := protocol.DecodeUpdate(payload)
			if errPayload != nil {
				errSend := session.SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
				)
				if errSend != nil {
					return
				}
				continue
			}

			var operationEnvelope protocol.OperationEnvelope

			errDecode := json.Unmarshal(
				update.Operation,
				&operationEnvelope,
			)
			if errDecode != nil {
				errSend := session.SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
				)
				if errSend != nil {
					return
				}
				continue
			}

			formattedType := strings.TrimSpace(operationEnvelope.Type)
			if formattedType == "" {
				errSend := session.SendErrorMessage(
					protocol.InvalidPayload,
					protocol.InvalidPayloadMessage,
				)
				if errSend != nil {
					return
				}
				continue
			}

			if formattedType == protocol.OpSnapshot {
				snapshot, errSnapshot := protocol.DecodeSnapshot(update)
				if errSnapshot != nil {
					errSend := session.SendErrorMessage(protocol.InvalidPayload, protocol.InvalidPayloadMessage)
					if errSend != nil {
						return
					}
					continue
				}
				errDelta := rs.storage.SaveSnapshotRoom(session.roomID, *snapshot.Delta)
				if errDelta != nil {
					errSend := session.SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage)
					if errSend != nil {
						return
					}
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
					errSend := session.SendErrorMessage(
						protocol.NotJoined,
						protocol.NotJoinedMessage,
					)
					if errSend != nil {
						return
					}
				} else {
					errSend := session.SendErrorMessage(
						protocol.InternalError,
						protocol.InternalErrorMessage,
					)
					if errSend != nil {
						return
					}
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
							errSend := session.SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage)
							if errSend != nil {
								return
							}
							continue
						}

						if errSendSync2 := session.Send(sync2); errSendSync2 != nil {
							return
						}
						continue
					}
					errSend := session.SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage)
					if errSend != nil {
						return
					}
					continue
				}

				sync2, errSync2 := internalSync.EncodeSyncStep2(delta)
				if errSync2 != nil {
					errSend := session.SendErrorMessage(protocol.InternalError, protocol.InternalErrorMessage)
					if errSend != nil {
						return
					}
					continue
				}

				if errSendSync2 := session.Send(sync2); errSendSync2 != nil {
					return
				}
			}

			continue

		default:
			errSend := session.SendErrorMessage(protocol.UnknownMessageCode, protocol.UnknownMessage)
			if errSend != nil {
				return
			}
			continue
		}
	}
}
