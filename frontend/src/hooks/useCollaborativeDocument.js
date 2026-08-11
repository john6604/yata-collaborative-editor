import { useState, useRef } from "react";
import { Document } from "../crdt/document";
import { encodeJoin } from "../protocol/operationsEncoder";
import { decodeIncomingMessage } from "../protocol/envelopeDecoder";
import { generateSync1, generateSync2 } from "../sync/vector";
import { encodeEnvelope } from "../protocol/envelopeEncoder";
import { DeleteOperation, InsertOperation, SnapshotOperation, Sync1Operation, Sync2Operation } from "../crdt/operations";

export function useCollaborativeDocument(documentID, displayName) {
    const [connectionStatus, setConnectionStatus] = useState("disconnected");
    const [content, setContent] = useState("");

    const document = useRef(null);
    const websocket = useRef(null);
    const pendingOperations = useRef([]);

    if (document.current === null) {
        document.current = new Document();
    }

    return {
        connectionStatus,
        content,
        document,
        websocket,
        pendingOperations
    };
}

function connect(documentID, displayName) {

    if (websocket.current === null || websocket.current.readyState === WebSocket.CLOSED) {
        const ws = new WebSocket("ws://localhost:8181/ws");
        websocket.current = ws;
        setConnectionStatus("connecting");
        ws.onopen = () => {
            const joinMessage = encodeJoin(documentID, displayName);
            const json = JSON.stringify(joinMessage);

            ws.send(json);
        };

        ws.onmessage = (event) => {
            const [messageType, operation] = decodeIncomingMessage(event.data);
            if (messageType === "join_ack") {
                setConnectionStatus("syncing");
                const sync1Operation = generateSync1(document.current);
                const sync1Message = encodeEnvelope(sync1Operation);
                const jsonSync1 = JSON.stringify(sync1Message);

                ws.send(jsonSync1);
            }

            if (messageType === "update" && operation instanceof Sync1Operation) {
                const sync2Operation = generateSync2(document.current, operation);
                const sync2Message = encodeEnvelope(sync2Operation);
                const jsonSync2 = JSON.stringify(sync2Message);

                ws.send(jsonSync2);
            } else if (messageType === "update" && operation instanceof Sync2Operation) {
                document.current.integrateDelta(operation.delta);
                setConnectionStatus("online");
                setContent(document.current.visibleContent());
            } else if (messageType === "update" && operation instanceof SnapshotOperation) {
                document.current.integrateDelta(operation.delta);
                setConnectionStatus("online");  
                setContent(document.current.visibleContent()); 
            }

            if (messageType === "update" && operation instanceof InsertOperation) {
                document.current.remoteInsert(operation.new_id, operation.origin_id, operation.right_id, operation.content);
                setContent(document.current.visibleContent());
            } else if (messageType === "update" && operation instanceof DeleteOperation) {
                document.current.remoteDelete(operation.target_id);
                setContent(document.current.visibleContent());
            }
        }
    }
}