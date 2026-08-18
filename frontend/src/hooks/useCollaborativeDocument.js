import { useState, useRef, useEffect, useLayoutEffect } from "react";
import { Document } from "../crdt/document";
import { encodeJoin } from "../protocol/operationsEncoder";
import { decodeIncomingMessage } from "../protocol/envelopeDecoder";
import { generateSnapshot, generateSync1, generateSync2 } from "../sync/vector";
import { encodeEnvelope } from "../protocol/envelopeEncoder";
import { DeleteOperation, InsertOperation, SnapshotOperation, Sync1Operation, Sync2Operation } from "../crdt/operations";

export function useCollaborativeDocument(documentID, displayName, editorRef) {
    const [connectionStatus, setConnectionStatus] = useState("disconnected");
    const [content, setContent] = useState("");

    const document = useRef(null);
    const websocket = useRef(null);
    const pendingOperations = useRef([]);
    const pendingSelection = useRef(null);
    const beforeInputState = useRef(null);
    const connectionStatusRef = useRef("disconnected");

    if (document.current === null) {
        document.current = new Document();
    }

    function updateConnectionStatus(status) {
        connectionStatusRef.current = status;
        setConnectionStatus(status);
    }

    function connect(documentID, displayName) {

        if (websocket.current === null || websocket.current.readyState === WebSocket.CLOSED) {
            const ws = new WebSocket("ws://localhost:8181/ws");
            websocket.current = ws;
            updateConnectionStatus("connecting");
            ws.onopen = () => {
                const joinMessage = encodeJoin(documentID, displayName);
                const json = JSON.stringify(joinMessage);

                ws.send(json);
            };

            ws.onmessage = (event) => {
                const [messageType, operation] = decodeIncomingMessage(event.data);
                
                let start;
                let end;

                if (messageType === "join_ack") {
                    updateConnectionStatus("syncing");
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
                    updateConnectionStatus("online");
                    flushPendingOperations();
                    const value = document.current.visibleContent();
                    if (value !== editorRef.current.value) {
                        pendingSelection.current = {
                            start: editorRef.current.selectionStart,
                            end: editorRef.current.selectionEnd,
                        }
                        setContent(document.current.visibleContent());
                    }
                } else if (messageType === "update" && operation instanceof SnapshotOperation) {
                    document.current.integrateDelta(operation.delta);
                    updateConnectionStatus("online");  
                    flushPendingOperations();
                    const value = document.current.visibleContent();
                    if (value !== editorRef.current.value) {
                        pendingSelection.current = {
                            start: editorRef.current.selectionStart,
                            end: editorRef.current.selectionEnd,
                        }
                        setContent(document.current.visibleContent());
                    }
                }

                if (messageType === "update" && operation instanceof InsertOperation) {
                    let value;
                    value = editorRef.current.value;
                    start = editorRef.current.selectionStart;
                    end = editorRef.current.selectionEnd;
                    let startCrdt = selectionToCRDTIndex(value, start);
                    let endCrdt = selectionToCRDTIndex(value, end);
                    document.current.remoteInsert(operation.new_id, operation.origin_id, operation.right_id, operation.content);
                    const index = document.current.visibleElementIndex(operation.new_id);

                    if (index <= startCrdt) {
                        startCrdt++;
                    }

                    if (index <= endCrdt) {
                        endCrdt++;
                    }

                    console.log("REMOTE CRDT:", document.current.visibleContent());
                    value = document.current.visibleContent();
                    start = crdtIndexToSelection(value, startCrdt);
                    end = crdtIndexToSelection(value, endCrdt);
                    
                    pendingSelection.current = {
                        start: start,
                        end: end,
                    }
                    setContent(document.current.visibleContent());
                } else if (messageType === "update" && operation instanceof DeleteOperation) {
                    const index = document.current.visibleElementIndex(operation.target_id);
                    let value = editorRef.current.value;
                    start = editorRef.current.selectionStart;
                    end = editorRef.current.selectionEnd;
                    let startCrdt = selectionToCRDTIndex(value, start);
                    let endCrdt = selectionToCRDTIndex(value, end);
                    document.current.remoteDelete(operation.target_id);
                    
                    if (index < startCrdt) {
                        startCrdt--;
                    }

                    if (index < endCrdt) {
                        endCrdt--;
                    }

                    value = document.current.visibleContent();
                    start = crdtIndexToSelection(value, startCrdt);
                    end = crdtIndexToSelection(value, endCrdt);

                    pendingSelection.current = {
                        start: start,
                        end: end,
                    }
                    setContent(document.current.visibleContent());
                }
            }
        } 

        if (websocket.current.readyState === WebSocket.CONNECTING || websocket.current.readyState === WebSocket.OPEN) {
            websocket.current.onclose = () => {
                updateConnectionStatus("disconnected");
                console.log("client has been disconnected...");
                setTimeout(() => {
                    connect(documentID, displayName);
                }, 2000);
            }
        }
    }

    useEffect(() => {
        connect(documentID, displayName);
    }, []);

    useLayoutEffect(() => {
        if (pendingSelection.current === null || editorRef.current === null) {
            return;
        }

        editorRef.current.setSelectionRange(pendingSelection.current.start, pendingSelection.current.end);

        pendingSelection.current = null;
    }, [content]);

    useEffect(() => {
        if (editorRef.current === null) {
            return;
        }

        const textarea = editorRef.current;
        textarea.addEventListener("beforeinput", handleBeforeInput);

        return () => {
            textarea.removeEventListener("beforeinput", handleBeforeInput);
        };
    }, []);


    function handleInput(event) {
        const eventType = event.nativeEvent.inputType;
        const data = event.nativeEvent.data;
        const selectionStart = event.currentTarget.selectionStart;
        const text = event.currentTarget.value;
        let operationID;

        switch (eventType) {
            case "insertText":
                if (beforeInputState.current.start === beforeInputState.current.end) {
                    const insertIndex = selectionToCRDTIndex(text, selectionStart);
                    operationID = document.current.insertElement(insertIndex - 1, data);
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, true);
                    sendSnapshot();
                } else {
                    const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                    const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                    const count = beforeEndCrdt - beforeStartCrdt;
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.deleteElement(beforeStartCrdt);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, false);
                    }
                    setContent(document.current.visibleContent());
                    operationID = document.current.insertElement(beforeStartCrdt, data);
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, true);
                    sendSnapshot();
                }
            break;
            case "deleteContentBackward":
                const deleteBackwardIndex = selectionToCRDTIndex(text, selectionStart);
                if (beforeInputState.current.start === beforeInputState.current.end) {
                    operationID = document.current.deleteElement(deleteBackwardIndex);
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, false);
                    sendSnapshot();
                } else {
                    const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                    const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                    const count = beforeEndCrdt - beforeStartCrdt;
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.deleteElement(beforeStartCrdt);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, false);
                    }
                    setContent(document.current.visibleContent());
                    sendSnapshot();
                }
            break;
            case "deleteContentForward":
                const deleteForwardIndex = selectionToCRDTIndex(text, selectionStart);
                if (beforeInputState.current.start === beforeInputState.current.end) {
                    operationID = document.current.deleteElement(deleteForwardIndex);
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, false);
                    sendSnapshot();
                } else {
                    const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                    const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                    const count = beforeEndCrdt - beforeStartCrdt;
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.deleteElement(beforeStartCrdt);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, false);
                    }
                    setContent(document.current.visibleContent());
                    sendSnapshot();
                }
            break;
            case "insertLineBreak":
                if (beforeInputState.current.start === beforeInputState.current.end) {
                    const insertIndex = selectionToCRDTIndex(text, selectionStart);
                    operationID = document.current.insertElement(insertIndex - 1, '\n');
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, true);
                    sendSnapshot();
                } else {
                    const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                    const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                    const count = beforeEndCrdt - beforeStartCrdt;
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.deleteElement(beforeStartCrdt);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, false);
                    }
                    setContent(document.current.visibleContent());
                    operationID = document.current.insertElement(beforeStartCrdt, '\n');
                    setContent(document.current.visibleContent());
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, true);
                    sendSnapshot();
                }
            break;
            case "insertFromPaste":
                if (beforeInputState.current.start === beforeInputState.current.end) {
                    const count = Array.from(data).length;
                    const endIndex = selectionToCRDTIndex(text, selectionStart);
                    let startIndex = endIndex - count;
                    const characters = Array.from(data);
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.insertElement(startIndex, characters[i]);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, true);
                        startIndex++;
                    }
                    setContent(document.current.visibleContent());
                    sendSnapshot();
                } else {
                    const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                    const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                    const n = beforeEndCrdt - beforeStartCrdt;
                    for (let i = 0; i < n; i++) {
                        operationID = document.current.deleteElement(beforeStartCrdt);
                        setContent(document.current.visibleContent());
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, false);
                    }
                    const count = Array.from(data).length;
                    const endIndex = selectionToCRDTIndex(text, selectionStart);
                    let startIndex = endIndex - count;
                    const characters = Array.from(data);
                    for (let i = 0; i < count; i++) {
                        operationID = document.current.insertElement(startIndex, characters[i]);
                        console.log("CRDT:", document.current.visibleContent());
                        sendOrQueue(operationID, true);
                        startIndex++;
                    }
                    setContent(document.current.visibleContent());
                    sendSnapshot();
                }
            break;
            case "deleteByCut":
                const beforeStartCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.start);
                const beforeEndCrdt = selectionToCRDTIndex(beforeInputState.current.value, beforeInputState.current.end);
                const n = beforeEndCrdt - beforeStartCrdt;
                for (let i = 0; i < n; i++) {
                    operationID = document.current.deleteElement(beforeStartCrdt);
                    console.log("CRDT:", document.current.visibleContent());
                    sendOrQueue(operationID, false);
                }
                setContent(document.current.visibleContent());
                sendSnapshot();
            break;
        }
        console.log("inputType: ", event.nativeEvent.inputType);
        console.log("selectionStart: ", event.currentTarget.selectionStart);
        console.log("selectionEnd: ", event.currentTarget.selectionEnd);
        console.log("data: ", event.nativeEvent.data);
        console.log("value: ", event.currentTarget.value);
    }

    function handleCompositionEnd(event) {
        const data = event.data;
        const text = event.currentTarget.value;
        const selectionStart = event.currentTarget.selectionStart;
        
        if (Array.from(data).length !== 1) {
            return
        }

        const index = selectionToCRDTIndex(text, selectionStart);
        const operationID = document.current.insertElement(index - 1, data);
        setContent(document.current.visibleContent());

        sendOrQueue(operationID, true);
        sendSnapshot();

        return operationID;
    }

    function handleBeforeInput(event) {
        const value = event.currentTarget.value;
        const start = event.currentTarget.selectionStart;
        const end = event.currentTarget.selectionEnd;

        beforeInputState.current = {
            value: value,
            start: start,
            end: end,
        }

        console.log("before value: ", beforeInputState.current.value);
        console.log("before start: ", beforeInputState.current.start);
        console.log("before end: ", beforeInputState.current.end);
    }

    function selectionToCRDTIndex(text, selectionStart) {
        const selectedtext = text.slice(0, selectionStart);
        const characters = Array.from(selectedtext);
        return characters.length;
    }

    function crdtIndexToSelection(text, crdtIndex) {
        let index = 0;

        const value = Array.from(text);
        const characters = value.slice(0, crdtIndex);

        for (const char of characters) {
            index += char.length;
        }

        return index;
    }

    function sendOrQueue(operationID, insert) {
        if (websocket.current.readyState === WebSocket.OPEN && connectionStatusRef.current === "online") {
            let operation;
            if (insert) {
                operation = document.current.insertLog.get(operationID.toKey());
            } else {
                operation = document.current.deleteLog.get(operationID.toKey());
            }
            const operationEnvelope = encodeEnvelope(operation);
            const jsonOperation = JSON.stringify(operationEnvelope);
            websocket.current.send(jsonOperation);
            return
        }

        if (insert) {
            pendingOperations.current.push({
                operation: operationID,
                type: "insert"
            });
        } else {
            pendingOperations.current.push({
                operation: operationID,
                type: "delete"
            });
        }
    }

    function flushPendingOperations() {
        while (pendingOperations.current.length > 0 && websocket.current.readyState === WebSocket.OPEN && connectionStatusRef.current === "online") {
            const operation = pendingOperations.current.shift();

            if (operation.type === "insert") {
                const insertOperation = document.current.insertLog.get(operation.operation.toKey());
                const insertOperationEnvelope = encodeEnvelope(insertOperation);
                const jsonInsert = JSON.stringify(insertOperationEnvelope);
                websocket.current.send(jsonInsert);
            } else if (operation.type === "delete") {
                const deleteOperation = document.current.deleteLog.get(operation.operation.toKey());
                const deleteOperationEnvelope = encodeEnvelope(deleteOperation);
                const jsonDelete = JSON.stringify(deleteOperationEnvelope);
                websocket.current.send(jsonDelete);
            }
        }
        sendSnapshot();
    }

    function sendSnapshot() {
        if (websocket.current.readyState === WebSocket.OPEN && connectionStatusRef.current === "online") {
            const snapshot = generateSnapshot(document.current);
            const snapshotEnvelope = encodeEnvelope(snapshot);
            const jsonSnapshot = JSON.stringify(snapshotEnvelope);
            websocket.current.send(jsonSnapshot);
        }
    }

    return {
        connectionStatus,
        content,
        document,
        websocket,
        pendingOperations,
        handleInput,
        handleCompositionEnd,
    };
}

