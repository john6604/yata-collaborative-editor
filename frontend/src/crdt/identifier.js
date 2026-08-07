// compatible fields with JSON structure in Go
export  class ID {
    constructor(clientID, clock) {
        this.client_id = clientID;
        this.clock = clock;
    }

    toKey() {
        return `${this.client_id}:${this.clock}`;
    }
    
    compareIDs(id) {
        return(this.client_id === id.client_id && this.clock === id.clock);
    }
}

export const startNode = new ID("START", -1);
export const endNode = new ID("END", -2);

export function validateID(message) {

    if (message === null || message === undefined) {
        throw new Error("no message found.")
    }
    
    if (message.client_id === undefined) {
        throw new Error("no client found.");
    }

    if (message.clock === undefined) {
        throw new Error("no clock found.");
    }

    if (typeof message.client_id !== "string") {
        throw new Error("client must be a string.");
    }

    if (!Number.isInteger(message.clock)) {
        throw new Error("clock must be an integer.");
    }

    return new ID(message.client_id, message.clock);
}

