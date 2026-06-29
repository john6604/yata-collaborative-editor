package sync

import (
	"sort"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

type Vector struct {
	StateVectors map[string]int
	DeleteSet    map[string][]int
}

func validateClockContinuity(document document.Document, idClient string, maxClock int) bool {

	for i := 0; i <= maxClock; i++ {
		id := identifier.NewID(idClient, i)
		_, exists := document.ElementsByID[*id]

		if !exists {
			return false
		}
	}

	return true
}

func (v *Vector) GenerateStateVector(elements document.Document) {

	for k := range elements.ElementsByID {

		clientID := k.ClientID
		clock := k.Clock

		if clientID == "START" || clientID == "END" {
			continue
		}

		value, exists := v.StateVectors[clientID]

		if !validateClockContinuity(elements, clientID, clock) {
			continue
		}

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

func deleteExists(knownDeletes []int, clock int) bool {
	for _, delete := range knownDeletes {
		if delete == clock {
			return true
		}
	}

	return false
}

func ComputeDelta(localDocument document.Document, vector Vector) ([]identifier.ID, []identifier.ID) {

	missingInserts := []identifier.ID{}
	missingDeletes := []identifier.ID{}

	insertsKnown := vector.StateVectors
	deletesKnown := vector.DeleteSet

	for k, v := range localDocument.ElementsByID {

		clientID := k.ClientID
		clock := k.Clock

		if clientID == "START" || clientID == "END" {
			continue
		}

		value, exists := insertsKnown[clientID]
		value1, exists1 := deletesKnown[clientID]

		if !exists {
			missingInserts = append(missingInserts, k)
		} else {
			if clock > value {
				missingInserts = append(missingInserts, k)
			}
		}

		if v.IsDeleted {
			if !exists1 {
				missingDeletes = append(missingDeletes, k)
			} else {
				if !deleteExists(value1, clock) {
					missingDeletes = append(missingDeletes, k)
				}
			}
		}
	}

	for k := range localDocument.PendingDeletes {

		clientID := k.ClientID
		clock := k.Clock

		if clientID == "START" || clientID == "END" {
			continue
		}

		value, exists := deletesKnown[clientID]

		if !exists {
			missingDeletes = append(missingDeletes, k)
		} else {
			if !deleteExists(value, clock) {
				missingDeletes = append(missingDeletes, k)
			}
		}
	}

	return missingInserts, missingDeletes
}

func ComputeSerializedDelta(localDocument document.Document, missingInserts []identifier.ID, missingDeletes []identifier.ID) protocol.Delta {

	var inserts []*protocol.InsertOperation

	for _, v := range missingInserts {

		value, exists := localDocument.InsertLog[v]

		if exists {
			inserts = append(inserts, value)
		}

	}

	var deletes []*protocol.DeleteOperation

	for _, v := range missingDeletes {

		value, exists := localDocument.DeleteLog[v]

		if exists {
			deletes = append(deletes, value)
		}
	}

	return protocol.Delta{Inserts: inserts, Deletes: deletes}
}
