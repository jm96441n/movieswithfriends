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
	router.HandleFunc("/movies/{id}", a.MoviesShowHandler).Methods("GET")
	router.HandleFunc("/movies/create", a.MoviesCreateHandler).Methods("POST")

	// parties related routes
	router.HandleFunc("/parties/{id}", a.PartyShowHandler).Methods("GET")
	router.HandleFunc("/parties/{id}", a.AddMovietoPartyHandler).Methods("PUT")

	// profiles related routes
	router.HandleFunc("/profiles/{id}", a.ProfileShowHandler).Methods("GET")
	return router
}
