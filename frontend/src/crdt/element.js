export class Element {
    constructor(id, origin, left, right, content) {
        this.elementID = id;
        this.origin = origin;
        this.left = left;
        this.right = right;
        this.content = content;
        this.isDeleted = false;
    }
}