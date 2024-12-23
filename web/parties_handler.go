package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/partymgmt"
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
		a.clientError(w, r, http.StatusBadRequest, "Name is required for creating a party")
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
	templateData := a.NewPartiesTemplateData(r, w, "/parties")
	a.render(w, r, http.StatusOK, "parties/new.gohtml", templateData)
}

func (a *Application) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "PartyShowHandler")
	ctx := r.Context()
	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("failed to get party ID from path", slog.Any("error", err))
		a.clientError(w, r, http.StatusBadRequest, "Failed to find the party, please try again")
		return
	}

	party, err := a.PartyService.GetPartyWithMovies(ctx, id)
	if err != nil {
		logger.Error("failed to get party by ID", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewPartiesTemplateData(r, w, "/parties")
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

	idAddedBy, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.Logger.Error("failed to get profile ID from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	idMovie, err := strconv.Atoi(r.FormValue("id_movie"))
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "Uh oh")
		return
	}

	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = a.PartiesRepository.AddMovieToParty(ctx, idParty, idMovie, idAddedBy)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	result, err := a.PartiesRepository.GetPartyByID(ctx, idParty)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	party := partymgmt.Party{
		ID:      result.ID,
		Name:    result.Name,
		ShortID: result.ShortID,
	}

	templateData := a.NewPartiesTemplateData(r, w, "/parties")
	templateData.Party = party

	a.renderPartial(w, r, http.StatusOK, "movies/partials/party_list_item.gohtml", templateData)
}

func (a *Application) MarkMovieAsWatchedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idPartyParam := r.PathValue("party_id")
	idMovieParam := r.PathValue("id")
	idMovie, err := strconv.Atoi(idMovieParam)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = a.PartiesRepository.MarkMovieAsWatched(ctx, idParty, idMovie)
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
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = a.PartiesRepository.SelectMovieForParty(ctx, idParty)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}
