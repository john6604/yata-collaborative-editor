package document

var startID = NewID("START", -1)
var endID = NewID("END", -2)

type Document struct {
	Start            *Element
	End              *Element
	ElementsByID     map[ID]*Element
	CharacterCounter int
	clientID         string
	clock            int
}
