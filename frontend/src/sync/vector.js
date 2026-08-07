import { ID } from "../crdt/identifier";
import { Sync1Operation, Sync2Operation } from "../crdt/operations";

export class Vector {
    constructor(stateVector, deleteSet) {
        this.StateVector = stateVector;
        this.DeleteSet = deleteSet;
    }

    generateStateVector(elements) {
        this.StateVector = new Map();

        for (const [key, value] of elements.insertLog.entries()) {
            
            const [clientID, auxClock] = key.split(":");
            const clock = Number(auxClock);

            if (clientID === "START" || clientID === "END") {
                continue
            }

            const exists = this.StateVector.has(clientID)

            if (!validateClockContinuity(elements, clientID, clock)) {
                continue
            }

            if (!exists) {
                this.StateVector.set(clientID, clock);
            } else {
                const value = this.StateVector.get(clientID);
                if (clock > value) {
                    this.StateVector.set(clientID, clock);
                }
            }
        }
    }

    generateDeleteSet(deletes) {
        this.DeleteSet = new Map();

        for (const [key, value] of deletes.deleteLog.entries()) {
            const [clientID, auxClock] = key.split(":");
            const clock = Number(auxClock);

            if (!this.DeleteSet.has(clientID)) {
                this.DeleteSet.set(clientID, []);
            }

            this.DeleteSet.get(clientID).push(clock);
        }

        const sortElements = sortClocks(this.DeleteSet);

        this.DeleteSet = sortElements;
    }
}

function validateClockContinuity(document, clientID, maxClock) {
    for (let i = 0; i <= maxClock; i++) {
        const id = new ID(clientID, i);
        const exists = document.insertLog.has(id.toKey());

        if (!exists) {
            return false
        }
    }

    return true
}

function sortClocks(clocks) {

    for (const [key, value] of clocks.entries()) {
        value.sort((a, b) => a - b);
    }

    return clocks;
}

function deleteExists(knownDeletes, clock) {
    for (const value of knownDeletes) {
        if (value === clock) {
            return true;
        }
    }

    return false;
}

export function computeDelta(localDocument, vector) {
    const missingInserts = [];
    const missingDeletes = [];

    const insertsKnown = vector.StateVector;
    const deletesKnown = vector.DeleteSet;

    for (const [key, value] of localDocument.insertLog.entries()) {
        const[clientID, auxClock] = key.split(":");
        const clock = Number(auxClock);

        if (clientID === "START" || clientID === "END") {
            continue;
        }

        const exists = insertsKnown.has(clientID);

        if (!exists) {
            missingInserts.push(new ID(clientID, clock));
        } else {
            const insert = insertsKnown.get(clientID);
            if (clock > insert) {
                missingInserts.push(new ID(clientID, clock));
            }
        }
    }

    for (const [key, value] of localDocument.deleteLog.entries()) {
        const[clientID, auxClock] = key.split(":");
        const clock = Number(auxClock);

        if (clientID === "START" || clientID === "END") {
            continue;
        }

        const exists = deletesKnown.has(clientID);

        if (!exists) {
            missingDeletes.push(new ID(clientID, clock));
        } else {
            const deletion = deletesKnown.get(clientID);
            if (!deleteExists(deletion, clock)) {
                missingDeletes.push(new ID(clientID, clock));
            }
        }
    }

    return [missingInserts, missingDeletes];
}

export function computeSerializedDelta(localDocument, missingInserts, missingDeletes) {
    const inserts = [];

    for (const value of missingInserts) {
        const exists = localDocument.insertLog.has(value.toKey());

        if (exists) {
            inserts.push(localDocument.insertLog.get(value.toKey()));
        }
    }

    const deletes = [];

    for (const value of missingDeletes) {
        const exists = localDocument.deleteLog.has(value.toKey());

        if (exists) {
            deletes.push(localDocument.deleteLog.get(value.toKey()));
        }
    }

    const delta = {
        inserts: inserts,
        deletes: deletes,
    }

    return delta;
}

export function generateSync1(document) {
    const vector = new Vector(null, null);

    vector.generateStateVector(document);
    vector.generateDeleteSet(document);

    return new Sync1Operation(vector.StateVector, vector.DeleteSet);
}

export function generateSync2(document, sync1Operation) {

    const vector = new Vector(sync1Operation.vector_state, sync1Operation.delete_set);

    const [missingInserts, missingDeletes] = computeDelta(document, vector);
    const delta = computeSerializedDelta(document, missingInserts, missingDeletes);

    return new Sync2Operation(delta);
}