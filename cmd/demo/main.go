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
	err = doc.InsertElement(100, 'X')
	fmt.Println(err)
	err = doc.Delete(-1)
	fmt.Println(err)
}
