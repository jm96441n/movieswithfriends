package web

import (
	"net/http"
	"strconv"
)

func (a *Application) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParam := r.PathValue("id")
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
	templateData := a.NewPartiesTemplateData(r, "/parties")
	templateData.Party = party
	a.render(w, r, http.StatusOK, "parties/show.gohtml", templateData)
}

func (a *Application) AddMovietoPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idPartyParam := r.PathValue("id")
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

	templateData := a.NewPartiesTemplateData(r, "/parties")
	templateData.Party = party

	a.renderPartial(w, r, http.StatusOK, "movies/partials/party_list_item.gohtml", templateData)
}

func (a *Application) MarkMovieAsWatchedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idPartyParam := r.PathValue("party_id")
	idMovieParam := r.PathValue("id")
	idMovie, err := strconv.Atoi(idMovieParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	err = a.PartiesService.MarkMovieAsWatched(ctx, idParty, idMovie)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}

func (a *Application) SelectMovieForParty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idPartyParam := r.PathValue("party_id")
	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	err = a.PartiesService.SelectMovieForParty(ctx, idParty)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}
