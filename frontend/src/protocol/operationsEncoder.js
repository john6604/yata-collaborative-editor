import { InsertOperation, DeleteOperation } from "../crdt/operations";

export function encodeInsertion(message) {
    const msg = {
        type: "insert",
        new_id: message.new_id,
        origin_id: message.origin_id,
        right_id: message.right_id,
        character: message.content,
    }

    return msg;
}

export function encodeDeletion(message) {
    const msg = {
        type: "delete",
        target_id: message.target_id,
    }

    return msg;
}

export function encodeUpdate(message) {

    let operationKind;

    if (message instanceof InsertOperation) {
        operationKind = "insert";
    } else if (message instanceof DeleteOperation) {
        operationKind = "delete";
    }

    let operation

    switch (operationKind) {
        case "insert":
            operation = encodeInsertion(message);
            break;
        case "delete":
            operation = encodeDeletion(message);
            break;
        default:
            throw new Error("unsupported operation");
    }

    return operation
}

export function encodeJoin(roomID, clientID) {
    
    if (typeof roomID !== "string" || typeof clientID !== "string") {
        throw new Error("field must be a string");
    }
    
    const room = roomID.trim();
    const client = clientID.trim();

    if (room === "" || client === "") {
        throw new Error("empty field")
    }

    const payload = {
        room: room,
        client_id: client,
    }

    const msg = {
        version: 1,
        type: "join",
        payload: payload,
    }

    return msg;
}