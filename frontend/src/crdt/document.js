import { Element } from "./element"
import { ID, startNode, endNode } from "./identifier"
import { InsertOperation, DeleteOperation } from "./operations";
import { PendingElement } from "./buffer";

export class Document {
    constructor() {
        const startDocument = new Element(startNode, null, null, null, '\x00');
        const endDocument = new Element(endNode, null, null, null, '\x00');
        this.elementsByID = new Map();
        this.characterCounter = 0;
        this.clientID = crypto.randomUUID();
        this.clock = 0;
        this.pendingInserts = new Map();
        this.pendingDeletes = new Map();
        this.insertLog = new Map();
        this.deleteLog = new Map();

        // Initial Logic 
        startDocument.right = endDocument;
        endDocument.left = startDocument;
        this.start = startDocument;
        this.end = endDocument;

        // Initial links
        this.elementsByID.set(startDocument.elementID.toKey(), startDocument);
        this.elementsByID.set(endDocument.elementID.toKey(), endDocument);
    }

    generateID() {
        const id = new ID(this.clientID, this.clock);
        this.clock++;
        return id;
    }

    findVisiblePosition(index) {
        if (index > this.characterCounter || index < 0) {
            throw new Error("Inexisting position to insert character."); 
        }

        let current = this.start.right;
        let visibleIndex = 0;

        while (current !== null) {
            if (!current.isDeleted) {
                if (visibleIndex === index) {
                    return [current.left, current];
                }
                visibleIndex++;
            }
            current = current.right
        }

        return [this.end.left, this.end]
    }

    findVisibleElement(index) {
        if (index >= this.characterCounter || index < 0) {
            throw new Error("The character to delete does not exist.");
        }

        let current = this.start.right;
        let visiblePosition = 0;

        while (current !== null) {
            if (!current.isDeleted) {
                if (visiblePosition === index) {
                    return current;
                }
                visiblePosition++;
            }
            current = current.right;
        }

        throw new Error("An unexpected error has occured.");
    }

    insertElement(index, character) {
        
        if (Array.from(character).length !== 1) {
            throw new Error("Must be exactly one character.");
        }

        const [previousElement, nextElement] = this.findVisiblePosition(index);

        const id = this.generateID();
        const insertedElement = new Element(id, previousElement, previousElement, nextElement, character);
        previousElement.right = insertedElement;
        nextElement.left = insertedElement;
        this.elementsByID.set(insertedElement.elementID.toKey(), insertedElement);
        const insertOperation = new InsertOperation(insertedElement.elementID, previousElement.elementID, nextElement.elementID, character);
        this.insertLog.set(insertedElement.elementID.toKey(), insertOperation);

        this.characterCounter++;


        return id
    }

    deleteElement(index) {
        const element = this.findVisibleElement(index);

        element.isDeleted = true;
        this.characterCounter--;
        const deleteOperation = new DeleteOperation(element.elementID);
        this.deleteLog.set(element.elementID.toKey(), deleteOperation);

        return element.elementID;
    }

    visibleContent() {
        const content = [];

        let current = this.start.right;

        while (current !== this.end) {
            if (!current.isDeleted) {
                content.push(current.content);
            }
            current = current.right;
        }

        return content.join("");
    }

    conflictZone(originID, rightID) {
        const origin = this.elementsByID.get(originID.toKey());
        const right = this.elementsByID.get(rightID.toKey());

        const conflictingElements = [];

        let current = origin.right;

        while (current !== right) {
            conflictingElements.push(current)
            current = current.right;
        }

        return conflictingElements;
    }

    isOriginAfter(insertOperation, conflictiveOperation) {
        let current = this.start.right;

        while (current !== this.end) {
            if (insertOperation === current) {
                return false
            }

            if (conflictiveOperation.origin === current) {
                return true
            }

            current = current.right
        }

        return false
    }

    findInsertionPoint(originID, rightID, newID) {
        const origin = this.elementsByID.get(originID.toKey());
        let left = origin
        let right = origin.right

        const conflictingElements = this.conflictZone(originID, rightID);

        for (const ops of conflictingElements) {
            if (this.isOriginAfter(origin, ops)) {
                break;
            }

            if (ops.elementID.clock < newID.clock) {
                left = ops;
                right = ops.right;
            } else if (ops.elementID.clock > newID.clock) {
                right = ops;
                break;
            } else {
                if (ops.elementID.client_id < newID.client_id) {
                    left = ops;
                    right = ops.right;
                } else {
                    right = ops;
                    break;
                }
            }
        }

        return [left, right];
    }

    processPending() {
        let progress = true;

        while (progress) {

            progress = false;

            for (const [key, value] of this.pendingInserts.entries()) {
                if (this.elementsByID.has(value.originID.toKey()) && this.elementsByID.has(value.rightID.toKey())) {
                    this.integrateInsert(value.newID, value.originID, value.rightID, value.content);
                    this.pendingInserts.delete(value.newID.toKey());
                    progress = true
                }
            }
        
        }
    }

    processPendingDeletes() {
        for (const [key, value] of this.pendingDeletes.entries()) {
            if (this.elementsByID.has(value.newID.toKey())) {
                this.integrateDeletion(value.newID);
                this.pendingDeletes.delete(value.newID.toKey());
            }
        }
    }

    integrateInsert(newID, originID, rightID, content) {
        
        let [left, right] = this.findInsertionPoint(originID, rightID, newID);
        const origin = this.elementsByID.get(originID.toKey());

        const element = new Element(newID, origin, left, right, content);

        left.right = element;
        right.left = element;

        this.elementsByID.set(newID.toKey(), element);
        const insertOperation = new InsertOperation(newID, originID, rightID, content);
        this.insertLog.set(newID.toKey(), insertOperation);

        this.characterCounter++;

    }

    remoteInsert(newID, originID, rightID, content) {
        if (this.elementsByID.has(newID.toKey())) {
            throw new Error("The value was already inserted.");
        }

        if (!this.elementsByID.has(originID.toKey()) || !this.elementsByID.has(rightID.toKey())) {
            this.pendingInserts.set(newID.toKey(), new PendingElement(newID, originID, rightID, content));
            const insertOperation = new InsertOperation(newID, originID, rightID, content);
            this.insertLog.set(newID.toKey(), insertOperation);
            throw new Error("Pending value.");
        }

        this.integrateInsert(newID, originID, rightID, content);

        this.processPending();

        this.processPendingDeletes(); 

    }

    integrateDeletion(elementID) {

        const element = this.elementsByID.get(elementID.toKey());
        element.isDeleted = true;

        this.characterCounter--;

        const deleteOperation = new DeleteOperation(element.elementID);
        this.deleteLog.set(element.elementID.toKey(), deleteOperation);
    }

    remoteDelete(elementID) {
        if (!this.elementsByID.has(elementID.toKey())) {
            this.pendingDeletes.set(elementID.toKey(), new PendingElement(elementID, null, null, '\x00'));
            const deleteOperation = new DeleteOperation(elementID);
            this.deleteLog.set(elementID.toKey(), deleteOperation);
            return
        }

        if (this.elementsByID.get(elementID.toKey()).isDeleted) {
            return
        }

        this.integrateDeletion(elementID);
    }
}