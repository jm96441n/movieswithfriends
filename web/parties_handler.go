package web

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *Application) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idParam := vars["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}
	party, err := a.PartiesService.GetPartyByIDWithMovies(ctx, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	templateData := a.NewMoviesTemplateData(r, "/parties")
	templateData.Party = party
	a.render(w, r, http.StatusOK, "parties/show.gohtml", templateData)
}

func (a *Application) AddMovietoPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idPartyParam := vars["id"]
	err := r.ParseForm()
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	idMovie, err := strconv.Atoi(r.FormValue("id_movie"))
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	err = a.PartiesService.AddMovieToParty(ctx, idParty, idMovie)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	party, err := a.PartiesService.GetPartyByID(ctx, idParty)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewMoviesTemplateData(r, "/parties")
	templateData.Party = party

	a.renderPartial(w, r, http.StatusOK, "movies/partials/party_list_item.gohtml", templateData)
}
