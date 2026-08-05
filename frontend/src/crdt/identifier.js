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

