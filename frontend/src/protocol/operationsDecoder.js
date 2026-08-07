import { InsertOperation, DeleteOperation, Sync1Operation, Sync2Operation, SnapshotOperation } from "../crdt/operations";
import { validateID } from "../crdt/identifier";

function isPlainObject(value) {
    return Object.prototype.toString.call(value) === "[object Object]";
}

function validateVectorState(vectorState) {
    if (!isPlainObject(vectorState)) {
        throw new Error("Expected object");
    }

    for (const [key, val] of Object.entries(vectorState)) {
        if (typeof key !== "string") {
            throw new Error("Key must be a string");
        }

        if (!Number.isInteger(val)) {
            throw new Error("Value must be an integer");
        }
    }
}

function validateGoInsertOperation(operation) {
    if (!isPlainObject(operation)) {
        throw new Error("Expected object");
    }

    if (operation.NewID === undefined || operation.NewID === null || operation.OriginID === undefined || operation.OriginID === null || operation.RightID === undefined || operation.RightID === null || operation.Content === undefined || operation.Content === null) {
        throw new Error("missing field");
    }

    validateID(operation.NewID);
    validateID(operation.OriginID);
    validateID(operation.RightID);

    if (!Number.isInteger(operation.Content)) {
        throw new Error("Content must be an integer");
    }

    if (operation.Content < 0 || operation.Content > 0x10FFFF) {
        throw new Error("Content must be a unicode code point");
    }
}

function validateGoDeleteOperation(operation) {
    if (!isPlainObject(operation)) {
        throw new Error("Expected object");
    }

    if (operation.TargetID === undefined || operation.TargetID === null) {
        throw new Error("missing field");
    }

    validateID(operation.TargetID);
}

function validateDelta(delta) {
    if (!isPlainObject(delta)) {
        throw new Error("Expected object");
    }

    const hasInserts = Object.prototype.hasOwnProperty.call(delta, "inserts");
    const hasDeletes = Object.prototype.hasOwnProperty.call(delta, "deletes");

    if (!hasInserts || !hasDeletes) {
        throw new Error("missing field");
    }

    if (Object.keys(delta).length !== 2) {
        throw new Error("invalid_payload");
    }

    if (delta.inserts !== null) {
        if (!Array.isArray(delta.inserts)) {
            throw new Error("Expected array");
        }

        for (const insert of delta.inserts) {
            validateGoInsertOperation(insert);
        }
    }

    if (delta.deletes !== null) {
        if (!Array.isArray(delta.deletes)) {
            throw new Error("Expected array");
        }

        for (const deletion of delta.deletes) {
            validateGoDeleteOperation(deletion);
        }
    }
}

export function decodeInsertion(message) {

    if (message === undefined || message === null) {
        throw new Error("invalid message");
    }

    if (message.type === undefined || message.new_id === undefined || message.origin_id === undefined || message.right_id === undefined || message.character === undefined || message.type === null || message.new_id === null || message.origin_id === null || message.right_id === null || message.character === null) {
        throw new Error("missing field");
    }

    if (message.type !== "insert") {
        throw new Error("unsupported operation");
    }

    if (typeof message.character !== "string") {
        throw new Error("character must be a string");
    }

    const newID = validateID(message.new_id);
    const originID = validateID(message.origin_id);
    const rightID = validateID(message.right_id);

    const characters = [...message.character];

    if (characters.length === 0) {
        throw new Error("missing_field");
    }

    if (characters.length !== 1) {
        throw new Error("invalid_payload");
    }

    return new InsertOperation(newID, originID, rightID, characters[0]);
}

export function decodeDeletion(message) {
    if (message === undefined || message === null) {
        throw new Error("invalid message");
    }

    if (message.type === undefined || message.type === null || message.target_id === undefined || message.target_id === null) {
        throw new Error("missing field");
    }

    if (message.type !== "delete") {
        throw new Error("unsupported operation");
    }

    const targetID = validateID(message.target_id);

    return new DeleteOperation(targetID);
}

export function decodeSync1(message) {
    if (message === undefined || message === null) {
        throw new Error("invalid message");
    }

    if (message.type === undefined || message.type === null || message.vector_state === undefined || message.vector_state === null || message.delete_set === undefined || message.delete_set === null) {
        throw new Error("missing field");
    }

    if (message.type !== "sync_step1") {
        throw new Error("unsupported operation");
    }

    validateVectorState(message.vector_state);

    if (!isPlainObject(message.delete_set)) {
        throw new Error("Expected object");
    }

    for (const value of Object.values(message.delete_set)) {
        if (!Array.isArray(value)) {
            throw new Error("Expected array");
        }

        for (const number of value) {
            if (!Number.isInteger(number)) {
                throw new Error("Expected integer");
            }
        }
    }

    const vectorState = new Map(Object.entries(message.vector_state));
    const deleteSet = new Map(Object.entries(message.delete_set));

    return new Sync1Operation(vectorState, deleteSet);
}

export function decodeSync2(message) {
    if (message === undefined || message === null) {
        throw new Error("invalid message");
    }

    if (message.type === undefined || message.type === null || message.delta === undefined || message.delta === null ) {
        throw new Error("missing field");
    }

    if (message.type !== "sync_step2") {
        throw new Error("unsupported operation");
    }

    const delta = decodeDelta(message.delta);

    return new Sync2Operation(delta);
}

export function decodeSnapshot(message) {
    if (message === undefined || message === null) {
        throw new Error("invalid message");
    }

    if (message.type === undefined || message.type === null || message.delta === undefined || message.delta === null ) {
        throw new Error("missing field");
    }

    if (message.type !== "snapshot") {
        throw new Error("unsupported operation");
    }

    const delta = decodeDelta(message.delta);

    return new SnapshotOperation(delta);
}

export function decodeDelta(delta) {
    validateDelta(delta);

    const inserts = [];
    const deletes = [];

    if (delta.inserts === undefined || delta.deletes === undefined) {
        throw new Error("delta must contain arrays");
    }

    for (const value of delta.inserts) {
        const newID = validateID(value.NewID);
        const originID = validateID(value.OriginID);
        const rightID = validateID(value.RightID);
        const content = String.fromCodePoint(value.Content);

        inserts.push(new InsertOperation(newID, originID, rightID, content));
    }

    for (const value of delta.deletes) {
        const targetID = validateID(value.TargetID);

        deletes.push(new DeleteOperation(targetID));
    }

    return {
        inserts: inserts,
        deletes: deletes,
    };
}

export function decodeUpdate(message) {
    if (message === undefined || message === null) {
        throw new Error("no message found");
    }

    if (message.type === undefined || message.type === null) {
        throw new Error("no operation found");
    }

    const type = message.type

    let operation;

    switch (type) {
        case "insert":
            operation = decodeInsertion(message);
            break;
        case "delete":
            operation = decodeDeletion(message);
            break;
        case "sync_step1":
            operation = decodeSync1(message);
            break;
        case "sync_step2":
            operation = decodeSync2(message);
            break;
        case "snapshot":
            operation = decodeSnapshot(message);
            break;
        default:
            throw new Error("unsupported operation");
    }

    return operation;
}

export function decodeJoinAck(payload) {
    if (payload === undefined || payload === null) {
        throw new Error("invalid payload");
    }

    if (payload.room === undefined || payload.room === null || payload.client_id === undefined || payload.client_id === null) {
        throw new Error("missing field");
    }

    if (typeof payload.room !== "string" || typeof payload.client_id !== "string") {
        throw new Error("field must be a string");
    }

    const room = payload.room.trim();
    const client = payload.client_id.trim();

    if (room === "" || client === "") {
        throw new Error("field must not be empty");
    }

    return [room, client];
}
