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