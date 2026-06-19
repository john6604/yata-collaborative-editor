package main

import (
	"fmt"

	"github.com/john6604/yata-collaborative-editor/internal/document"
)

func main() {

	//Test 1
	fmt.Println("TEST 1")
	doc := document.NewDocument()
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test 2
	fmt.Println("TEST 2")
	doc.InsertElement(0, 'H')
	doc.InsertElement(1, 'e')
	doc.InsertElement(2, 'l')
	doc.InsertElement(3, 'l')
	doc.InsertElement(4, 'o')
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test 3
	fmt.Println("TEST 3")
	doc.InsertElement(0, 'X')
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test 4
	fmt.Println("TEST 4")
	doc.InsertElement(3, 'Y')
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test 5
	fmt.Println("TEST 5")
	doc.Delete(0)
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test 6
	fmt.Println("TEST 6")
	doc.Delete(2)
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test7
	fmt.Println("TEST 7")
	doc.Delete(1)
	doc.InsertElement(1, 'a')
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test8
	fmt.Println("TEST 8")
	doc.Delete(4)
	fmt.Println(doc)
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())
	doc.InsertElement(4, 'o')
	fmt.Println(doc.VisibleLength())
	fmt.Println(doc.PrintInternal())

	//Test9
	fmt.Println("TEST 9")
	err := doc.Delete(100)
	fmt.Println(err)
	err, _ = doc.InsertElement(100, 'X')
	fmt.Println(err)
	err = doc.Delete(-1)
	fmt.Println(err)

	//Test Concurrencia 1
	fmt.Println("TEST CONCURRENCIA 1")
	var StartID = document.ID{

		ClientID: "START",

		Clock: -1,
	}
	docA := document.NewDocument()
	docB := document.NewDocument()
	docC := document.NewDocument()

	_, idA := docA.InsertElement(0, 'X')
	_, idB := docB.InsertElement(0, 'Y')

	//Propagation
	docB.RemoteInsert(StartID, idA, 'X')
	docC.RemoteInsert(StartID, idA, 'X')

	docA.RemoteInsert(StartID, idB, 'Y')
	docC.RemoteInsert(StartID, idB, 'Y')

	//Visualization
	fmt.Println("Replica A")
	fmt.Println(docA)
	fmt.Println("Replica B")
	fmt.Println(docB)
	fmt.Println("Replica C")
	fmt.Println(docC)

	// Test Concurrencia 2
	fmt.Println("TEST CONCURRENCIA 2")
	docA = document.NewDocument()
	docB = document.NewDocument()
	docC = document.NewDocument()

	_, idA = docA.InsertElement(0, 'X')

	docB.RemoteInsert(StartID, idA, 'X')
	docC.RemoteInsert(StartID, idA, 'X')

	docA.Delete(0)
	docB.RemoteDelete(idA)
	docC.RemoteDelete(idA)

	fmt.Println("Replica A")
	fmt.Println(docA)
	fmt.Println("Replica B")
	fmt.Println(docB)
	fmt.Println("Replica C")
	fmt.Println(docC)

	// Test Concurrencia 3

	fmt.Println("TEST CONCURRENCIA 3")
	docA = document.NewDocument()
	docB = document.NewDocument()
	docC = document.NewDocument()

	_, idA = docA.InsertElement(0, 'X')

	docB.RemoteInsert(StartID, idA, 'X')
	docB.RemoteInsert(StartID, idA, 'X')

	docA.Delete(0)

	docB.RemoteDelete(idA)
	docB.RemoteDelete(idA)

	fmt.Println("Replica A")
	fmt.Println(docA)
	fmt.Println("Replica B")
	fmt.Println(docB)

	// Test Concurrencia 4

	fmt.Println("TEST CONCURRENCIA 4")
	docA = document.NewDocument()
	docB = document.NewDocument()

	_, idA1 := docA.InsertElement(0, 'A')
	_, idA2 := docA.InsertElement(1, 'X')
	_, idA3 := docA.InsertElement(2, 'P')
	_, idA4 := docA.InsertElement(3, 'Y')

	docB.RemoteInsert(StartID, idA1, 'A')
	docB.RemoteInsert(StartID, idA2, 'X')
	docB.RemoteInsert(StartID, idA3, 'P')
	docB.RemoteInsert(StartID, idA4, 'Y')

	_, idA5 := docA.InsertElement(1, 'W')
	docB.RemoteInsert(idA1, idA5, 'W')

	fmt.Println("Replica A")
	fmt.Println(docA)
	fmt.Println("Replica B")
	fmt.Println(docB)
}
