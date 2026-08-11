import { decodeError, decodeJoinAck, decodeUpdate } from "./operationsDecoder";

export function decodeEnvelope(message) {

    const msg = JSON.parse(message);

    if (msg === undefined || msg === null) {
        throw new Error("no message found");
    }

    if (msg.version === undefined || msg.version === null || msg.type === undefined || msg.type === null || msg.payload === undefined || msg.payload === null) {
        throw new Error("missing field");
    }

    if (!Number.isInteger(msg.version)) {
        throw new Error("version must be an integer");
    }

    if (typeof msg.type !== "string") {
        throw new Error("type must be a string");
    }

    const type = msg.type.trim();

    if (msg.version !== 1) {
        throw new Error("unsupported version");
    }

    if (type === "") {
        throw new Error("empty type");
    }

    return msg;
}

export function decodeUpdateOperation(payload) {
    if (payload === undefined || payload === null) {
        throw new Error("invalid payload");
    }

    if (payload.op === undefined || payload.op === null) {
        throw new Error("missing field");
    }

    const operation = decodeUpdate(payload.op);

    return operation;
}

export function decodeIncomingMessage(message) {
    const msg = decodeEnvelope(message);
    let operation

    switch (msg.type) {
        case "join_ack":
            operation = decodeJoinAck(msg.payload);
            break;
        case "update":
            operation = decodeUpdateOperation(msg.payload);
            break;
        case "error":
            operation = decodeError(msg.payload);
            break;
        default:
            throw new Error("unsupported operation");
    }

    return [msg.type, operation];
}