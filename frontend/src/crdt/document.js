import { Element } from "./element"
import { ID, startNode, endNode } from "./identifier"
import { InsertOperation, DeleteOperation } from "./operations";

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
}