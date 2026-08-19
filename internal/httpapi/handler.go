package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type Value struct {
	Name string `json:"name"`
}

func (api *API) documents(response http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case http.MethodGet:
		list, err := api.store.ListDocuments()
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

		documents, errJson := json.Marshal(list)
		if errJson != nil {
			http.Error(response, errJson.Error(), http.StatusInternalServerError)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.Write(documents)
	case http.MethodPost:
		var name Value

		err := json.NewDecoder(request.Body).Decode(&name)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		if name.Name == "" {
			http.Error(response, "name cannot be empty", http.StatusBadRequest)
			return
		}

		id, errUUID := generateUUID()
		if errUUID != nil {
			http.Error(response, errUUID.Error(), http.StatusInternalServerError)
			return
		}

		documentMetada, errCreate := api.store.CreateDocument(id, name.Name)
		if errCreate != nil {
			http.Error(response, "could not create the document", http.StatusInternalServerError)
			return
		}

		data, errJson := json.Marshal(documentMetada)
		if errJson != nil {
			http.Error(response, "could not create the document", http.StatusInternalServerError)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusCreated)
		response.Write(data)
	default:
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (api *API) patchDocument(response http.ResponseWriter, request *http.Request) {

	if request.Method != http.MethodPatch {
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := request.PathValue("id")
	var name Value

	err := json.NewDecoder(request.Body).Decode(&name)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	if name.Name == "" {
		http.Error(response, "name cannot be empty", http.StatusBadRequest)
		return
	}

	errRename := api.store.RenameDocument(id, name.Name)
	if errRename != nil {
		http.Error(response, errRename.Error(), http.StatusInternalServerError)
		return
	}

	doc, errGet := api.store.GetDocumentByID(id)
	if errGet != nil {
		if errGet == sql.ErrNoRows {
			http.Error(response, errGet.Error(), http.StatusNotFound)
			return
		}
		http.Error(response, errGet.Error(), http.StatusInternalServerError)
		return
	}

	data, errJson := json.Marshal(doc)
	if errJson != nil {
		http.Error(response, errJson.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(data)
}

// Function to generate a random UUID
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}
