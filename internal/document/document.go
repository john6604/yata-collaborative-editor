package document

import (
	"crypto/rand"
	"errors"
	"fmt"
)

// Package variables for Start and End nodes
var startID = NewID("START", -1)
var endID = NewID("END", -2)

// Definition os structure
type Document struct {
	Start            *Element
	End              *Element
	ElementsByID     map[ID]*Element
	CharacterCounter int
	clientID         string
	clock            int
}

// Generate a random UUID
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

// Generate a new document
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

// Generate an ID for an element
func (d *Document) generateElementID() *ID {
	id := NewID(d.clientID, d.clock)
	d.clock++
	return id
}

// Finding visible position
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

// Local insertion in the document
func (d *Document) insertElement(index int, character byte) error {

	elements := make(map[int]*Element)
	previousElement := NewEmptyElement()
	nextElement := NewEmptyElement()

	if index > d.CharacterCounter {
		return errors.New("Inexisting position to insert character.")
	}

	for k := range d.ElementsByID {
		i := 0
		if d.ElementsByID[k].IsDeleted == false {
			elements[i] = d.ElementsByID[k]
			i++
		}
	}

	for k := range elements {
		if k == index && k == 0 {
			previousElement = d.Start
			nextElement = elements[k+1]
		} else if k == index && k+1 == -2 {
			previousElement = elements[k-1]
			nextElement = d.End
		} else {
			previousElement = elements[k-1]
			nextElement = elements[k+1]
		}
	}

	id := d.generateElementID()
	insertedElement := NewElement(*id, previousElement, previousElement, nextElement, character)
	d.ElementsByID[insertedElement.ElementID] = insertedElement

	d.CharacterCounter++

	return nil
}
