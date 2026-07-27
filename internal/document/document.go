package document

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

// Package variables for Start and End nodes
var startID = identifier.NewID("START", -1)
var endID = identifier.NewID("END", -2)

// Definition of structure
type Document struct {
	Start            *Element
	End              *Element
	ElementsByID     map[identifier.ID]*Element
	CharacterCounter int
	ClientID         string
	Clock            int
	PendingInserts   map[identifier.ID]*PendingElement
	PendingDeletes   map[identifier.ID]*PendingElement
	InsertLog        map[identifier.ID]*protocol.InsertOperation
	DeleteLog        map[identifier.ID]*protocol.DeleteOperation
}

// Function to generate a new document
func NewDocument() *Document {

	//Document attributes defined
	document := Document{CharacterCounter: 0}
	start := NewElement(*startID, nil, nil, nil, '\x00')
	end := NewElement(*endID, nil, nil, nil, '\x00')
	document.ElementsByID = make(map[identifier.ID]*Element)
	document.PendingInserts = make(map[identifier.ID]*PendingElement)
	document.PendingDeletes = make(map[identifier.ID]*PendingElement)
	document.InsertLog = make(map[identifier.ID]*protocol.InsertOperation)
	document.DeleteLog = make(map[identifier.ID]*protocol.DeleteOperation)
	document.CharacterCounter = 0
	document.ClientID = generateUUID()
	document.Clock = 0

	//Initial logic defined
	start.Right = end
	end.Left = start
	document.Start = start
	document.End = end

	//Initialize linked list
	document.ElementsByID[start.ElementID] = start
	document.ElementsByID[end.ElementID] = end

	return &document
}

// Function to generate a random UUID
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err.Error()
	}

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

// Function to generate an ID for an element
func (d *Document) generateElementID() *identifier.ID {
	id := identifier.NewID(d.ClientID, d.Clock)
	d.Clock++
	return id
}

// Function to insert locally in the document
func (d *Document) InsertElement(index int, character rune) (error, identifier.ID) {

	previousElement, nextElement, err := d.findVisiblePosition(index)

	if err != nil {
		return err, identifier.ID{}
	}

	id := d.generateElementID()
	insertedElement := NewElement(*id, previousElement, previousElement, nextElement, character)
	previousElement.Right = insertedElement
	nextElement.Left = insertedElement
	d.ElementsByID[insertedElement.ElementID] = insertedElement
	insertOperation := protocol.NewInsertOperation(insertedElement.ElementID, previousElement.ElementID, nextElement.ElementID, character)
	d.InsertLog[insertedElement.ElementID] = insertOperation

	d.CharacterCounter++

	return nil, *id
}

func (d *Document) integrateInsert(newID identifier.ID, originID identifier.ID, rightID identifier.ID, content rune) {

	left, right := d.findInsetionPoint(originID, rightID, newID)
	origin := d.ElementsByID[originID]

	element := NewElement(newID, origin, left, right, content)

	left.Right = element
	right.Left = element

	d.ElementsByID[newID] = element
	insertOperation := protocol.NewInsertOperation(newID, originID, rightID, content)
	d.InsertLog[newID] = insertOperation

	d.CharacterCounter++

}

// Function to insert remotely/concurrently in the document
func (d *Document) RemoteInsert(newID identifier.ID, originID identifier.ID, rightID identifier.ID, content rune) error {

	if d.ElementsByID[newID] != nil {
		return errors.New("The value was already inserted.")
	}

	if d.ElementsByID[originID] == nil || d.ElementsByID[rightID] == nil {
		d.PendingInserts[newID] = NewPending(newID, originID, rightID, content)
		insertOperation := protocol.NewInsertOperation(newID, originID, rightID, content)
		d.InsertLog[newID] = insertOperation
		return errors.New("Pending value.")
	}

	d.integrateInsert(newID, originID, rightID, content)

	d.processPending()

	d.processPendingDeletes()

	return nil
}

// Function to delete an element logically
func (d *Document) Delete(index int) (error, identifier.ID) {

	element, err := d.findVisibleElement(index)

	if err != nil {
		return err, identifier.ID{}
	}

	element.IsDeleted = true
	d.CharacterCounter--
	deleteOperation := protocol.NewDeleteOperation(element.ElementID)
	d.DeleteLog[element.ElementID] = deleteOperation

	return nil, element.ElementID
}

func (d *Document) integrateDeletion(elementID identifier.ID) {

	element := d.ElementsByID[elementID]
	element.IsDeleted = true

	d.CharacterCounter--

	deleteOperation := protocol.NewDeleteOperation(element.ElementID)
	d.DeleteLog[element.ElementID] = deleteOperation
}

// Function to delete an element remotely
func (d *Document) RemoteDelete(elementID identifier.ID) error {

	if d.ElementsByID[elementID] == nil {
		d.PendingDeletes[elementID] = NewPending(elementID, identifier.ID{}, identifier.ID{}, '\x00')
		deleteOperation := protocol.NewDeleteOperation(elementID)
		d.DeleteLog[elementID] = deleteOperation
		return nil
	}

	if d.ElementsByID[elementID].IsDeleted == true {
		return nil
	}

	d.integrateDeletion(elementID)

	return nil
}

func (d *Document) IntegrateDelta(delta protocol.Delta) error {

	var inserts []*protocol.InsertOperation
	inserts = delta.Inserts

	for _, v := range inserts {
		err := d.RemoteInsert(v.NewID, v.OriginID, v.RightID, v.Content)
		if err != nil {
			if err.Error() != "Pending value." && err.Error() != "The value was already inserted." {
				return err
			}
		}
	}

	var deletes []*protocol.DeleteOperation
	deletes = delta.Deletes

	for _, v := range deletes {
		err := d.RemoteDelete(v.TargetID)
		if err != nil {
			return err
		}
	}

	return nil
}
