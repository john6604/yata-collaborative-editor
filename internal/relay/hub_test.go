package relay_test

/*
import (
	"fmt"
	"sync"
	"testing"

	"github.com/john6604/yata-collaborative-editor/internal/relay"
)

func TestNewHubStartsEmpty(t *testing.T) {

	hub := relay.NewHub()

	got := hub.CountRooms()
	expected := 0

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}
}

func TestJoinCreatesRoomAndRegistersClient(t *testing.T) {

	hub := relay.NewHub()

	session, err := hub.Join("room-1", "client-A")

	if err != nil {
		t.Fatalf("An error %s occurred", err)
	}

	if session == nil {
		t.Fatal("An error occurred.")
	}

	got := hub.CountRooms()
	expected := 1

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	clients, err1 := hub.CountClients("room-1")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	got1 := clients
	expected1 := 1

	if got1 != expected1 {
		t.Fatalf(
			"%d clients were expected, but got %d clients.",
			expected1,
			got1,
		)
	}

	hasClient, err2 := hub.HasClient("room-1", "client-A")

	if err2 != nil {
		t.Fatalf("An error %s occurred", err2)
	}

	expectedFlag := true

	if hasClient != expectedFlag {
		t.Fatalf(
			"%t was expected, but got %t.",
			expectedFlag,
			hasClient,
		)
	}
}

func TestJoinRegistersTwoClientsInSameRoom(t *testing.T) {

	hub := relay.NewHub()

	session1, err := hub.Join("room-1", "client-A")

	if err != nil {
		t.Fatalf("An error %s occurred", err)
	}

	if session1 == nil {
		t.Fatal("An error occurred.")
	}

	session2, err1 := hub.Join("room-1", "client-B")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	if session2 == nil {
		t.Fatal("An error occurred.")
	}

	got := hub.CountRooms()
	expected := 1

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	clients, err2 := hub.CountClients("room-1")

	if err2 != nil {
		t.Fatalf("An error %s occurred", err2)
	}

	got1 := clients
	expected1 := 2

	if got1 != expected1 {
		t.Fatalf(
			"%d clients were expected, but got %d clients.",
			expected1,
			got1,
		)
	}

	hasClientA, err4 := hub.HasClient("room-1", "client-A")

	if err4 != nil {
		t.Fatalf("An error %s occurred", err4)
	}

	expectedFlagA := true

	if hasClientA != expectedFlagA {
		t.Fatalf(
			"%t was expected, but got %t.",
			expectedFlagA,
			hasClientA,
		)
	}

	hasClientB, err5 := hub.HasClient("room-1", "client-B")

	if err5 != nil {
		t.Fatalf("An error %s occurred", err5)
	}

	expectedFlagB := true

	if hasClientB != expectedFlagB {
		t.Fatalf(
			"%t was expected, but got %t.",
			expectedFlagB,
			hasClientB,
		)
	}
}

func TestJoinRejectsDuplicateClient(t *testing.T) {

	hub := relay.NewHub()

	session1, err1 := hub.Join("room-1", "client-A")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	if session1 == nil {
		t.Fatal("An error occurred.")
	}

	session2, err2 := hub.Join("room-1", "client-A")

	if err2 == nil {
		t.Fatalf("An error %s occurred", err2)
	}

	if session2 != nil {
		t.Fatal("An error occurred.")
	}

	got := hub.CountRooms()
	expected := 1

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	clients, err3 := hub.CountClients("room-1")

	if err3 != nil {
		t.Fatalf("An error %s occurred", err3)
	}

	got1 := clients
	expected1 := 1

	if got1 != expected1 {
		t.Fatalf(
			"%d clients were expected, but got %d clients.",
			expected1,
			got1,
		)
	}

	hasClient, err4 := hub.HasClient("room-1", "client-A")

	if err4 != nil {
		t.Fatalf("An error %s occurred", err4)
	}

	expectedFlag := true

	if hasClient != expectedFlag {
		t.Fatalf(
			"%t was expected, but got %t.",
			expectedFlag,
			hasClient,
		)
	}

}

func TestSameClientIDCanJoinDifferentRooms(t *testing.T) {

	hub := relay.NewHub()

	session1, err1 := hub.Join("room-1", "client-A")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	if session1 == nil {
		t.Fatal("An error occurred.")
	}

	session2, err2 := hub.Join("room-2", "client-A")

	if err2 != nil {
		t.Fatalf("An error %s occurred", err2)
	}

	if session2 == nil {
		t.Fatal("An error occurred.")
	}

	got := hub.CountRooms()
	expected := 2

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	hasClientInRoom1, err3 := hub.HasClient("room-1", "client-A")

	if err3 != nil {
		t.Fatalf("An error %s occurred", err3)
	}

	if !hasClientInRoom1 {
		t.Fatal("client-A was expected in room-1.")
	}

	hasClientInRoom2, err4 := hub.HasClient("room-2", "client-A")

	if err4 != nil {
		t.Fatalf("An error %s occurred", err4)
	}

	if !hasClientInRoom2 {
		t.Fatal("client-A was expected in room-2.")
	}
}

func TestLeaveRemovesOnlySelectedClient(t *testing.T) {

	hub := relay.NewHub()

	session1, err1 := hub.Join("room-1", "client-A")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	_, err2 := hub.Join("room-1", "client-B")

	if err2 != nil {
		t.Fatalf("An error %s occurred", err2)
	}

	hub.Leave(session1)

	got := hub.CountRooms()
	expected := 1

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	clients, err3 := hub.CountClients("room-1")

	if err3 != nil {
		t.Fatalf("An error %s occurred", err3)
	}

	expectedClients := 1

	if clients != expectedClients {
		t.Fatalf(
			"%d clients were expected, but got %d clients.",
			expectedClients,
			clients,
		)
	}

	hasClientA, err4 := hub.HasClient("room-1", "client-A")

	if err4 != nil {
		t.Fatalf("An error %s occurred", err4)
	}

	if hasClientA {
		t.Fatal("client-A was not expected in room-1.")
	}

	hasClientB, err5 := hub.HasClient("room-1", "client-B")

	if err5 != nil {
		t.Fatalf("An error %s occurred", err5)
	}

	if !hasClientB {
		t.Fatal("client-B was expected in room-1.")
	}
}

func TestLeaveLastClientRemovesRoom(t *testing.T) {

	hub := relay.NewHub()

	session, err := hub.Join("room-1", "client-A")

	if err != nil {
		t.Fatalf("An error %s occurred", err)
	}

	hub.Leave(session)

	got := hub.CountRooms()
	expected := 0

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	_, err1 := hub.CountClients("room-1")

	if err1 == nil {
		t.Fatal("An error was expected.")
	}
}

func TestLeaveTwiceDoesNotPanic(t *testing.T) {

	hub := relay.NewHub()

	session, err := hub.Join("room-1", "client-A")

	if err != nil {
		t.Fatalf("An error %s occurred", err)
	}

	hub.Leave(session)
	hub.Leave(session)

	got := hub.CountRooms()
	expected := 0

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}
}

func TestLeaveNilDoesNotPanic(t *testing.T) {

	hub := relay.NewHub()

	hub.Leave(nil)

	got := hub.CountRooms()
	expected := 0

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}
}

func TestOldSessionDoesNotRemoveNewSession(t *testing.T) {

	hub := relay.NewHub()

	oldSession, err1 := hub.Join("room-1", "client-A")

	if err1 != nil {
		t.Fatalf("An error %s occurred", err1)
	}

	hub.Leave(oldSession)

	newSession, err2 := hub.Join("room-1", "client-A")

	if err2 != nil {
		t.Fatalf("An error %s occurred", err2)
	}

	if newSession == nil {
		t.Fatal("An error occurred.")
	}

	hub.Leave(oldSession)

	got := hub.CountRooms()
	expected := 1

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}

	clients, err3 := hub.CountClients("room-1")

	if err3 != nil {
		t.Fatalf("An error %s occurred", err3)
	}

	expectedClients := 1

	if clients != expectedClients {
		t.Fatalf(
			"%d clients were expected, but got %d clients.",
			expectedClients,
			clients,
		)
	}

	hasClient, err4 := hub.HasClient("room-1", "client-A")

	if err4 != nil {
		t.Fatalf("An error %s occurred", err4)
	}

	if !hasClient {
		t.Fatal("client-A was expected in room-1.")
	}
}

func TestJoinRejectsEmptyRoomIDAndClientID(t *testing.T) {

	tests := []struct {
		name     string
		roomID   string
		clientID string
	}{
		{name: "empty room id", roomID: "", clientID: "client-A"},
		{name: "blank room id", roomID: " ", clientID: "client-A"},
		{name: "empty client id", roomID: "room-1", clientID: ""},
		{name: "blank client id", roomID: "room-1", clientID: " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hub := relay.NewHub()

			session, err := hub.Join(tt.roomID, tt.clientID)

			if err == nil {
				t.Fatal("An error was expected.")
			}

			if session != nil {
				t.Fatal("A nil session was expected.")
			}

			got := hub.CountRooms()
			expected := 0

			if got != expected {
				t.Fatalf(
					"%d were expected, but got %d rooms.",
					expected,
					got,
				)
			}
		})
	}
}

func TestHubConcurrentJoinLeaveAndRead(t *testing.T) {

	hub := relay.NewHub()
	const goroutines = 50

	start := make(chan struct{})
	errs := make(chan error, goroutines)
	var waitGroup sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		waitGroup.Add(1)

		go func(index int) {
			defer waitGroup.Done()
			<-start

			roomID := fmt.Sprintf("room-%d", index%5)
			clientID := fmt.Sprintf("client-%d", index)

			session, err := hub.Join(roomID, clientID)

			if err != nil {
				errs <- err
				return
			}

			if session == nil {
				errs <- fmt.Errorf("session was nil")
				return
			}

			_, countErr := hub.CountClients(roomID)

			if countErr != nil {
				errs <- countErr
				return
			}

			hasClient, hasClientErr := hub.HasClient(roomID, clientID)

			if hasClientErr != nil {
				errs <- hasClientErr
				return
			}

			if !hasClient {
				errs <- fmt.Errorf("%s was expected in %s", clientID, roomID)
				return
			}

			hub.Leave(session)
		}(i)
	}

	close(start)
	waitGroup.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("An error %s occurred", err)
	}

	got := hub.CountRooms()
	expected := 0

	if got != expected {
		t.Fatalf(
			"%d were expected, but got %d rooms.",
			expected,
			got,
		)
	}
}
*/
