export class InsertOperation {
    constructor(newID, originID, rightID, content) {
        this.new_id = newID;
        this.origin_id = originID;
        this.right_id = rightID;
        this.content = content;
    }
}

export class DeleteOperation {
    constructor(targetID) {
        this.target_id = targetID;
    }
}

export class Sync1Operation {
    constructor(vectorState, deleteSet) {
        this.vector_state = vectorState;
        this.delete_set = deleteSet;
    }
}

export class Sync2Operation {
    constructor(delta) {
        this.delta = delta;
    }
}

export class SnapshotOperation {
    constructor(delta) {
        this.delta = delta;
    }
}

export class PresenceOperation {
    constructor(users) {
        this.users = users;
    }
}