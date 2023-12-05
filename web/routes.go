package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Application) Routes() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", a.HomeHandler)
	return router
}
