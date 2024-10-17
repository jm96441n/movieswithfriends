package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

func (a *Application) CreatePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.Logger.Error("failed to get profile ID from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		a.Logger.Error("failed to get parseForm", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		a.Logger.Error("failed to get name from form")
		a.clientError(w, http.StatusBadRequest)
		return
	}

	id, err := a.PartyService.CreateParty(ctx, profileID, name)
	if err != nil {
		a.Logger.Error("failed to get create party", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/parties/%d", id), http.StatusSeeOther)
}

func (a *Application) NewPartyHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "NewPartyHandler")
	logger.Info("calling NewPartyHandler")
	templateData := a.NewPartiesTemplateData(r, "/parties")
	a.render(w, r, http.StatusOK, "parties/new.gohtml", templateData)
}

func (a *Application) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "PartyShowHandler")
	ctx := r.Context()
	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("failed to get party ID from path", slog.Any("error", err))
		a.clientError(w, http.StatusBadRequest)
		return
	}

	party, err := a.PartiesStoreService.GetPartyByIDWithMovies(ctx, id)
	if err != nil {
		logger.Error("failed to get party by ID with movies", slog.Any("error", err))
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

	err = a.PartiesStoreService.AddMovieToParty(ctx, idParty, idMovie)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	party, err := a.PartiesStoreService.GetPartyByID(ctx, idParty)
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

	err = a.PartiesStoreService.MarkMovieAsWatched(ctx, idParty, idMovie)
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

	err = a.PartiesStoreService.SelectMovieForParty(ctx, idParty)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}
