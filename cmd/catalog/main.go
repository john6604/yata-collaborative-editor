package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/john6604/yata-collaborative-editor/internal/catalog"
)

func defaultDatabasePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("data", "catalog.db")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "data", "catalog.db"))
}

func main() {
	store, err := catalog.NewStore(defaultDatabasePath())
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	_, errCreate := store.CreateDocument("document-1", "Default Document")
	if errCreate != nil {
		log.Fatal(errCreate)
	}

	list, errList := store.ListDocuments()
	if errList != nil {
		log.Fatal(errList)
	}

	for _, doc := range list {
		fmt.Println(doc)
	}

	fmt.Println("Rename document: ")

	if errRename := store.RenameDocument("document-1", "Document 1"); errRename != nil {
		log.Fatal(errRename)
	}

	list2, errList2 := store.ListDocuments()
	if errList2 != nil {
		log.Fatal(errList2)
	}

	for _, doc := range list2 {
		fmt.Println(doc)
	}

}
