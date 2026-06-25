package sync

import (
	"sort"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
)

type Vector struct {
	StateVectors map[string]int
	DeleteSet    map[string][]int
}

func (v *Vector) GenerateStateVector(elements document.Document) {

	for k := range elements.ElementsByID {

		clientID := k.ClientID
		clock := k.Clock

		if clientID == "START" || clientID == "END" {
			continue
		}

		value, exists := v.StateVectors[clientID]

		if !exists {
			v.StateVectors[clientID] = clock
		} else {
			if clock > value {
				v.StateVectors[clientID] = clock
			}
		}

	}

}

func sortClocks(clocks map[string][]int) map[string][]int {

	for _, v := range clocks {
		sort.Ints(v)
	}

	return clocks
}

func (v *Vector) GenerateDeleteSet(deletes document.Document) {

	v.DeleteSet = make(map[string][]int)

	for k, value := range deletes.ElementsByID {

		clientID := k.ClientID
		clock := k.Clock

		if value.IsDeleted == true {
			v.DeleteSet[clientID] = append(v.DeleteSet[clientID], clock)
		}

	}

	for k := range deletes.PendingDeletes {

		clientID := k.ClientID
		clock := k.Clock

		v.DeleteSet[clientID] = append(v.DeleteSet[clientID], clock)

	}

	sortElements := sortClocks(v.DeleteSet)

	v.DeleteSet = sortElements
}

func ComputeDelta(localDocument document.Document, vector Vector) ([]identifier.ID, []identifier.ID) {

	missingInserts := []identifier.ID{}
	missingDeletes := []identifier.ID{}

	insertsKnown := vector.StateVectors
	deletesKnown := vector.DeleteSet

	for k, v := range localDocument.ElementsByID {

		clientID := k.ClientID
		clock := k.Clock

		id := identifier.NewID(clientID, clock)

		if clientID == "START" || clientID == "END" {
			continue
		}

		value, exists := insertsKnown[clientID]

		if !exists {
			missingInserts[*id] = document.NewPending(*id, v.Origin.ElementID, v.Right.ElementID, v.Content)
		} else {
			if insertsKnown[clientID] < clock {
				missingInserts[*id] = 
			}
		}
	}

	return missingInserts, missingDeletes
}
