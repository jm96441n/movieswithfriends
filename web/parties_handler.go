package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

func (a *Application) NewPartyHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "NewPartyHandler")
	logger.Info("calling NewPartyHandler")
	templateData := a.NewPartiesTemplateData(r, w, "/parties")
	a.render(w, r, http.StatusOK, "parties/new.gohtml", templateData)
}

func (a *Application) EditPartyHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "EditPartyHandler")
	logger.Info("calling EditPartyHandler")
	templateData := a.NewPartiesTemplateData(r, w, "/parties")
	a.render(w, r, http.StatusOK, "parties/edit.gohtml", templateData)
}

func (a *Application) CreatePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.Logger.Error("failed to get profile ID from session", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := a.NewPartiesTemplateData(r, w, "/parties")
		a.render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	err = r.ParseForm()
	if err != nil {
		a.Logger.Error("failed to get parseForm", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := a.NewPartiesTemplateData(r, w, "/parties")
		a.render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	name := r.FormValue("partyName")
	if name == "" {
		a.Logger.Error("failed to get partyName from form")
		a.setErrorFlashMessage(w, r, "A Name is required to create this party.")
		data := a.NewPartiesTemplateData(r, w, "/parties")
		a.render(w, r, http.StatusBadRequest, "parties/new.gohtml", data)
		return
	}

	id, err := a.PartyService.CreateParty(ctx, profileID, name)
	if err != nil {
		a.Logger.Error("failed to create party", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := a.NewPartiesTemplateData(r, w, "/parties")
		a.render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	a.setInfoFlashMessage(w, r, "Party successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/parties/%d", id), http.StatusSeeOther)
}

func (a *Application) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "PartyShowHandler")
	ctx := r.Context()

	watcher, err := a.getWatcherFromSession(r)
	if err != nil {
		logger.Error("failed to get party ID from path", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusInternalServerError)
		return
	}

	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.Error("failed to get party ID from path", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusBadRequest)
		return
	}

	party, err := a.PartyService.GetPartyWithMovies(ctx, logger, id)
	if err != nil {
		logger.Error("failed to get party by ID", slog.Any("error", err))
		data := a.NewTemplateData(r, w, "/parties")
		a.render(w, r, http.StatusNotFound, "404.gohtml", data)
		return
	}

	currentWatcherIsOwner, err := watcher.IsOwnerOfParty(ctx, id)
	if err != nil {
		logger.Error("failed to get if current user is owner", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusBadRequest)
		return
	}

	invites, err := a.InvitationsService.GetInvitationsForParty(ctx, id)
	if err != nil {
		logger.Error("failed to get invitations", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusBadRequest)
		return
	}

	templateData := a.NewPartiesTemplateData(r, w, "/parties")
	templateData.Party = party
	templateData.ModalData.PendingInvites = invites
	templateData.ModalData.PartyID = id
	templateData.CurrentWatcherIsOwner = currentWatcherIsOwner

	a.render(w, r, http.StatusOK, "parties/show.gohtml", templateData)
}

func (a *Application) PartiesIndexHandler(w http.ResponseWriter, r *http.Request) {
	// we set parties in the template creation
	// this will need to change when we use htmx to just replace the body
	// and not touch the sidebar
	templateData := a.NewPartiesTemplateData(r, w, "/parties")

	a.render(w, r, http.StatusOK, "parties/index.gohtml", templateData)
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

	err = a.PartiesRepository.MarkPartyMovieAsWatched(ctx, idParty, idMovie)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}

func (a *Application) SelectMovieForParty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "SelectMovieForParty")
	idPartyParam := r.PathValue("party_id")
	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		logger.Error("failed to get party ID from path", slog.Any("error", err))
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = a.PartiesRepository.SelectMovieForParty(ctx, idParty)
	if err != nil {
		logger.Error("failed to select movie for party", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}
