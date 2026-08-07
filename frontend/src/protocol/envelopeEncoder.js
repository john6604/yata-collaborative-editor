import { encodeUpdate } from "./operationsEncoder"

export function encodeEnvelope(message) {

    const operation =  encodeUpdate(message);
    const op = {
        op: operation,
    }

    const msg = {
        version: 1,
        type: "update",
        payload: op,
    }

    return msg;
}