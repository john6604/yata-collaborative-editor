package document

import (
	"errors"
	"strings"
)

// Function to find visible position (Insertion)
func (d *Document) findVisiblePosition(index int) (*Element, *Element, error) {

	if index > d.CharacterCounter || index < 0 {
		return nil, nil, errors.New("Inexisting position to insert character.")
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

// Function to find an element by its index (Deletion)
func (d *Document) findVisibleElement(index int) (*Element, error) {

	if index >= d.CharacterCounter || index < 0 {
		return nil, errors.New("The character to delete does not exist.")
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

// Function to visualize the current content in the document
func (d *Document) VisibleContent() string {

	var content strings.Builder

	current := d.Start.Right

	for current != d.End {
		if !current.IsDeleted {
			content.WriteString(string(current.Content))
		}
		current = current.Right
	}

	return content.String()
}

// Function to call directly to Println
func (d *Document) String() string {
	return d.VisibleContent()
}

// Function to print internal status of the document
func (d *Document) PrintInternal() string {

	var content strings.Builder
	content.WriteString(string(d.Start.ElementID.ClientID))

	current := d.Start.Right

	for current != d.End {
		content.WriteString(" -> ")
		if current.IsDeleted {
			content.WriteString(string(current.Content))
			content.WriteString("(X)")
		} else {
			content.WriteString(string(current.Content))
		}

		current = current.Right
	}

	content.WriteString(string(" -> "))
	content.WriteString(string(d.End.ElementID.ClientID))

	return content.String()
}

// Function to observe the character counter
func (d *Document) VisibleLength() int {
	return d.CharacterCounter
}

// Function to return an array of the elements in the list
func (d *Document) Traverse() []*Element {

	var elements []*Element
	current := d.Start

	for current != nil {
		elements = append(elements, current)
		current = current.Right
	}

	return elements
}
