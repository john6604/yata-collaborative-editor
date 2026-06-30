package sync

import (
	"reflect"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/document"
	"github.com/john6604/yata-collaborative-editor/internal/identifier"
	"github.com/john6604/yata-collaborative-editor/internal/protocol"
)

var (
	syncStartID = identifier.ID{ClientID: "START", Clock: -1}
	syncEndID   = identifier.ID{ClientID: "END", Clock: -2}
)

func addInsertLog(doc *document.Document, id, originID, rightID identifier.ID, content byte) {
	doc.InsertLog[id] = protocol.NewInsertOperation(id, originID, rightID, content)
}

func idSet(ids []identifier.ID) map[identifier.ID]struct{} {
	set := make(map[identifier.ID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func TestGenerateStateVectorUsesInsertLogAndContiguousClocks(t *testing.T) {
	doc := document.NewDocument()

	a0 := identifier.ID{ClientID: "client-A", Clock: 0}
	a1 := identifier.ID{ClientID: "client-A", Clock: 1}
	a2 := identifier.ID{ClientID: "client-A", Clock: 2}
	b0 := identifier.ID{ClientID: "client-B", Clock: 0}
	b2 := identifier.ID{ClientID: "client-B", Clock: 2}
	c1 := identifier.ID{ClientID: "client-C", Clock: 1}

	addInsertLog(doc, a0, syncStartID, syncEndID, 'A')
	addInsertLog(doc, a1, a0, syncEndID, 'B')
	addInsertLog(doc, a2, a1, syncEndID, 'C')
	addInsertLog(doc, b0, syncStartID, syncEndID, 'X')
	addInsertLog(doc, b2, b0, syncEndID, 'Z')          // client-B clock 1 is missing.
	addInsertLog(doc, c1, syncStartID, syncEndID, 'Q') // client-C clock 0 is missing.

	// None of these operations is present in ElementsByID. The vector must be
	// derived from InsertLog, including operations that are still pending.
	for _, id := range []identifier.ID{a0, a1, a2, b0, b2, c1} {
		if _, exists := doc.ElementsByID[id]; exists {
			t.Fatalf("test setup error: %v unexpectedly exists in ElementsByID", id)
		}
	}

	vector := Vector{}
	vector.GenerateStateVector(*doc)

	want := map[string]int{
		"client-A": 2,
		"client-B": 0,
	}
	if !reflect.DeepEqual(vector.StateVectors, want) {
		t.Errorf("StateVectors = %#v; expected %#v", vector.StateVectors, want)
	}
	if _, exists := vector.StateVectors["client-C"]; exists {
		t.Errorf("client-C must not be advertised because clock 0 is missing")
	}
}

func TestGenerateDeleteSetUsesDeleteLogGroupsAndSorts(t *testing.T) {
	doc := document.NewDocument()

	a2 := identifier.ID{ClientID: "client-A", Clock: 2}
	a0 := identifier.ID{ClientID: "client-A", Clock: 0}
	b3 := identifier.ID{ClientID: "client-B", Clock: 3}

	doc.DeleteLog[a2] = protocol.NewDeleteOperation(a2)
	doc.DeleteLog[a0] = protocol.NewDeleteOperation(a0)
	doc.DeleteLog[b3] = protocol.NewDeleteOperation(b3)

	// The targets are intentionally absent from the list. DeleteSet must still
	// expose deletes known through DeleteLog.
	for _, id := range []identifier.ID{a2, a0, b3} {
		if _, exists := doc.ElementsByID[id]; exists {
			t.Fatalf("test setup error: %v unexpectedly exists in ElementsByID", id)
		}
	}

	vector := Vector{}
	vector.GenerateDeleteSet(*doc)

	want := map[string][]int{
		"client-A": {0, 2},
		"client-B": {3},
	}
	if !reflect.DeepEqual(vector.DeleteSet, want) {
		t.Errorf("DeleteSet = %#v; expected %#v", vector.DeleteSet, want)
	}
}

func TestComputeDeltaReturnsExactlyMissingInserts(t *testing.T) {
	local := document.NewDocument()

	a0 := identifier.ID{ClientID: "client-A", Clock: 0}
	a1 := identifier.ID{ClientID: "client-A", Clock: 1}
	a2 := identifier.ID{ClientID: "client-A", Clock: 2}
	b0 := identifier.ID{ClientID: "client-B", Clock: 0}
	b1 := identifier.ID{ClientID: "client-B", Clock: 1}
	c0 := identifier.ID{ClientID: "client-C", Clock: 0}

	for _, id := range []identifier.ID{a0, a1, a2, b0, b1, c0} {
		addInsertLog(local, id, syncStartID, syncEndID, byte(id.Clock+'0'))
	}

	remote := Vector{
		StateVectors: map[string]int{
			"client-A": 0,
			"client-B": 1,
		},
		DeleteSet: map[string][]int{},
	}

	missingInserts, missingDeletes := ComputeDelta(*local, remote)

	want := idSet([]identifier.ID{a1, a2, c0})
	if got := idSet(missingInserts); !reflect.DeepEqual(got, want) {
		t.Errorf("missing inserts = %#v; expected %#v", got, want)
	}
	if len(missingDeletes) != 0 {
		t.Errorf("missing deletes = %#v; expected none", missingDeletes)
	}
}

func TestComputeDeltaReturnsOnlyUnknownDeletes(t *testing.T) {
	local := document.NewDocument()

	a0 := identifier.ID{ClientID: "client-A", Clock: 0}
	a2 := identifier.ID{ClientID: "client-A", Clock: 2}
	b1 := identifier.ID{ClientID: "client-B", Clock: 1}

	local.DeleteLog[a0] = protocol.NewDeleteOperation(a0)
	local.DeleteLog[a2] = protocol.NewDeleteOperation(a2)
	local.DeleteLog[b1] = protocol.NewDeleteOperation(b1)

	remote := Vector{
		StateVectors: map[string]int{
			"client-A": 2,
			"client-B": 1,
		},
		DeleteSet: map[string][]int{
			"client-A": {0},
			"client-B": {1},
		},
	}

	missingInserts, missingDeletes := ComputeDelta(*local, remote)

	if len(missingInserts) != 0 {
		t.Errorf("missing inserts = %#v; expected none", missingInserts)
	}
	want := idSet([]identifier.ID{a2})
	if got := idSet(missingDeletes); !reflect.DeepEqual(got, want) {
		t.Errorf("missing deletes = %#v; expected %#v", got, want)
	}
}

func TestComputeSerializedDeltaContainsExpectedOperations(t *testing.T) {
	local := document.NewDocument()

	a0 := identifier.ID{ClientID: "client-A", Clock: 0}
	a1 := identifier.ID{ClientID: "client-A", Clock: 1}
	b0 := identifier.ID{ClientID: "client-B", Clock: 0}

	addInsertLog(local, a0, syncStartID, syncEndID, 'A')
	addInsertLog(local, a1, a0, syncEndID, 'B')
	addInsertLog(local, b0, syncStartID, syncEndID, 'X')
	local.DeleteLog[a1] = protocol.NewDeleteOperation(a1)

	delta := ComputeSerializedDelta(
		*local,
		[]identifier.ID{a1, b0},
		[]identifier.ID{a1},
	)

	if got, want := len(delta.Inserts), 2; got != want {
		t.Fatalf("len(delta.Inserts) = %d; expected %d", got, want)
	}
	if got, want := len(delta.Deletes), 1; got != want {
		t.Fatalf("len(delta.Deletes) = %d; expected %d", got, want)
	}

	insertByID := make(map[identifier.ID]*protocol.InsertOperation)
	for _, operation := range delta.Inserts {
		insertByID[operation.NewID] = operation
	}

	if got := insertByID[a1]; got == nil {
		t.Fatalf("serialized delta does not contain insert %v", a1)
	} else {
		if got.OriginID != a0 {
			t.Errorf("insert %v OriginID = %v; expected %v", a1, got.OriginID, a0)
		}
		if got.RightID != syncEndID {
			t.Errorf("insert %v RightID = %v; expected %v", a1, got.RightID, syncEndID)
		}
		if got.Content != 'B' {
			t.Errorf("insert %v Content = %q; expected %q", a1, got.Content, 'B')
		}
	}

	if got := insertByID[b0]; got == nil {
		t.Fatalf("serialized delta does not contain insert %v", b0)
	} else if got.OriginID != syncStartID || got.RightID != syncEndID || got.Content != 'X' {
		t.Errorf("insert %v = %#v; expected START/END references and content X", b0, got)
	}

	if got := delta.Deletes[0].TargetID; got != a1 {
		t.Errorf("delete TargetID = %v; expected %v", got, a1)
	}
}
