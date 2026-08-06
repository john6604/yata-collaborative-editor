import { InsertOperation, DeleteOperation } from "../crdt/operations";
import { validateID } from "../crdt/identifier";

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