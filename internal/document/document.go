package document

import (
	"crypto/rand"
	"errors"
	"fmt"
)

// Package variables for Start and End nodes
var startID = NewID("START", -1)
var endID = NewID("END", -2)

// Definition of structure
type Document struct {
	Start            *Element
	End              *Element
	ElementsByID     map[ID]*Element
	CharacterCounter int
	clientID         string
	clock            int
	PendingInserts   map[ID]*PendingElement
	PendingDeletes   map[ID]*PendingElement
}

// Function to generate a new document
func NewDocument() *Document {

	//Document attributes defined
	document := Document{CharacterCounter: 0}
	start := NewElement(*startID, nil, nil, nil, '\x00')
	end := NewElement(*endID, nil, nil, nil, '\x00')
	document.ElementsByID = make(map[ID]*Element)
	document.PendingInserts = make(map[ID]*PendingElement)
	document.PendingDeletes = make(map[ID]*PendingElement)
	document.CharacterCounter = 0
	document.clientID = generateUUID()
	document.clock = 0

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
func (d *Document) generateElementID() *ID {
	id := NewID(d.clientID, d.clock)
	d.clock++
	return id
}

// Function to insert locally in the document
func (d *Document) InsertElement(index int, character byte) (error, ID) {

	previousElement, nextElement, err := d.findVisiblePosition(index)

	if err != nil {
		return err, ID{}
	}

	id := d.generateElementID()
	insertedElement := NewElement(*id, previousElement, previousElement, nextElement, character)
	previousElement.Right = insertedElement
	nextElement.Left = insertedElement
	d.ElementsByID[insertedElement.ElementID] = insertedElement

	d.CharacterCounter++

	return nil, *id
}

func (d *Document) integrateInsert(newID ID, originID ID, rightID ID, content byte) {

	left, right := d.findInsetionPoint(originID, rightID, newID)
	origin := d.ElementsByID[originID]

	element := NewElement(newID, origin, left, right, content)

	left.Right = element
	right.Left = element

	d.ElementsByID[newID] = element

	d.CharacterCounter++

}

// Function to insert remotely/concurrently in the document
func (d *Document) RemoteInsert(newID ID, originID ID, rightID ID, content byte) error {

	if d.ElementsByID[newID] != nil {
		return errors.New("The value was already inserted.")
	}

	if d.ElementsByID[originID] == nil || d.ElementsByID[rightID] == nil {
		d.PendingInserts[newID] = NewPending(newID, originID, rightID, content)
		return errors.New("Pending value.")
	}

	d.integrateInsert(newID, originID, rightID, content)

	d.processPendingDeletes()

	d.processPending()

	return nil
}

// Function to delete an element logically
func (d *Document) Delete(index int) error {

	element, err := d.findVisibleElement(index)

	if err != nil {
		return err
	}

	element.IsDeleted = true
	d.CharacterCounter--

	return nil
}

func (d *Document) integrateDeletion(elementID ID) {

	element := d.ElementsByID[elementID]
	element.IsDeleted = true

	d.CharacterCounter--
}

// Function to delete an element remotely
func (d *Document) RemoteDelete(elementID ID) error {

	if d.ElementsByID[elementID] == nil {
		d.PendingDeletes[elementID] = NewPending(elementID, ID{}, ID{}, '\x00')
		return nil
	}

	d.integrateDeletion(elementID)

	return nil
}
