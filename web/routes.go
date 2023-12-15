package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Application) Routes() http.Handler {
	router := mux.NewRouter()

	//	fileServer := http.FileServer(http.FS(ui.TemplateFS))
	//	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	router.HandleFunc("/", a.HomeHandler)

	// movies related routes
	router.HandleFunc("/movies", a.MoviesIndexHandler).Methods("GET")
	router.HandleFunc("/movies", a.MoviesSearchHandler).Methods("POST")
	router.HandleFunc("/movies/{id}", a.MoviesShowHandler).Methods("GET")
	return router
}
