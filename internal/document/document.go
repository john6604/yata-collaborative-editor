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

// Function to generate a new document
func NewDocument() *Document {

	//Document attributes defined
	document := Document{CharacterCounter: 0}
	start := NewElement(*startID, nil, nil, nil, '\x00')
	end := NewElement(*endID, nil, nil, nil, '\x00')
	document.ElementsByID = make(map[ID]*Element)
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

// Function to generate an ID for an element
func (d *Document) generateElementID() *ID {
	id := NewID(d.clientID, d.clock)
	d.clock++
	return id
}

// Function to find visible position (Insertion)
func (d *Document) findVisiblePosition(index int) (*Element, *Element, error) {

	if index > d.CharacterCounter || index < 0 {
		return nil, nil, errors.New("Inexisting position to insert character.")
	}

	if index == 0 {
		return d.Start, d.Start.Right, nil
	}

	current := d.Start.Right

	visibleIndex := 0

	for current != nil {
		if !current.IsDeleted {
			if visibleIndex == index {
				return current.Left, current, nil
			}
			visibleIndex++
		}
		current = current.Right
	}

	return d.End.Left, d.End, nil
}

// Function to insert locally in the document
func (d *Document) insertElement(index int, character byte) error {

	previousElement, nextElement, err := d.findVisiblePosition(index)

	if err != nil {
		return err
	}

	id := d.generateElementID()
	insertedElement := NewElement(*id, previousElement, previousElement, nextElement, character)
	previousElement.Right = insertedElement
	nextElement.Left = insertedElement
	d.ElementsByID[insertedElement.ElementID] = insertedElement

	d.CharacterCounter++

	return nil
}

// Function to find an element by its index (Deletion)
func (d *Document) findVisibleElement(index int) (*Element, error) {

	if index >= d.CharacterCounter || index < 0 {
		return nil, errors.New("The character to delete does not exist.")
	}

	if index == 0 {
		return d.Start.Right, nil
	}

	current := d.Start.Right

	visiblePosition := 0

	for current != nil {
		if !current.IsDeleted {
			if visiblePosition == index {
				return current, nil
			}
			visiblePosition++
		}
		current = current.Right
	}

	return nil, errors.New("An unexpected error has occured.")
}

// Function to delete an element logically
func (d *Document) delete(index int) error {

	element, err := d.findVisibleElement(index)

	if err != nil {
		return err
	}

	element.IsDeleted = true
	d.CharacterCounter--

	return nil
}
