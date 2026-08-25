package httpapi

import (
	"net/http"

	"github.com/john6604/yata-collaborative-editor/internal/catalog"
)

type API struct {
	store *catalog.Store
}

func NewAPI(store *catalog.Store) *API {
	api := API{
		store: store,
	}

	return &api
}

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/documents", api.documents)
	mux.HandleFunc("/api/documents/{id}", api.patchDocument)
}
