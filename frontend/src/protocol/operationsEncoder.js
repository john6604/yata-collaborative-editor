import { InsertOperation, DeleteOperation, Sync1Operation, Sync2Operation, SnapshotOperation } from "../crdt/operations";

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

export function encodeSync1(message) {
    
    const vectorState = Object.fromEntries(message.vector_state);
    const deleteSet = Object.fromEntries(message.delete_set);

    const msg = {
        type: "sync_step1",
        vector_state: vectorState,
        delete_set: deleteSet,
    }

    return msg;
}

export function encodeSync2(message) {
    
    const delta = encodeDelta(message.delta);
    
    const msg = {
        type: "sync_step2",
        delta: delta,
    }

    return msg;
}

export function encodeSnapshot(message) {
    
    const delta = encodeDelta(message.delta);
    
    const msg = {
        type: "snapshot",
        delta: delta,
    }

    return msg;
}

export function encodeDelta(delta) {

    const inserts = [];
    const deletes = [];

    if (delta === undefined || delta === null) {
        throw new Error("no delta found");
    }

    if (!Array.isArray(delta.inserts) || !Array.isArray(delta.deletes)) {
        throw new Error("delta must contain arrays");
    }

    for (const value of delta.inserts) {
        const newValue = {
            NewID: value.new_id,
            OriginID: value.origin_id,
            RightID: value.right_id,
            Content: value.content.codePointAt(0),
        }

        inserts.push(newValue);
    }

    for (const value of delta.deletes) {
        const newValue = {
            TargetID: value.target_id,
        }

        deletes.push(newValue);
    }

    const msg = {
        inserts: inserts,
        deletes: deletes,
    }

    return msg;
}

export function encodeUpdate(message) {

    let operationKind;

    if (message instanceof InsertOperation) {
        operationKind = "insert";
    } else if (message instanceof DeleteOperation) {
        operationKind = "delete";
    } else if (message instanceof Sync1Operation) {
        operationKind = "sync_step1";
    } else if (message instanceof Sync2Operation) {
        operationKind = "sync_step2";
    } else if (message instanceof SnapshotOperation) {
        operationKind = "snapshot";
    }

    let operation

    switch (operationKind) {
        case "insert":
            operation = encodeInsertion(message);
            break;
        case "delete":
            operation = encodeDeletion(message);
            break;
        case "sync_step1":
            operation = encodeSync1(message);
            break;
        case "sync_step2":
            operation = encodeSync2(message);
            break;
        case "snapshot":
            operation = encodeSnapshot(message);
            break;
        default:
            throw new Error("unsupported operation");
    }

    return operation
}

export function encodeJoin(roomID, clientID, name) {
    
    if (typeof roomID !== "string" || typeof clientID !== "string" || typeof name !== "string") {
        throw new Error("field must be a string");
    }
    
    const room = roomID.trim();
    const client = clientID.trim();
    const displayName = name.trim();

    if (room === "" || client === "" || displayName === "") {
        throw new Error("empty field")
    }

    const payload = {
        room: room,
        client_id: client,
        name: displayName,
    }

    const msg = {
        version: 1,
        type: "join",
        payload: payload,
    }

    return msg;
}