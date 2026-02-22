package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

// NewPartyHandler handles the new party form page.
func (h *Handlers) NewPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "NewPartyHandler")
	logger.InfoContext(ctx, "calling NewPartyHandler")
	templateData := h.NewPartiesTemplateData(r, w, "/parties")
	h.renderer.Render(w, r, http.StatusOK, "parties/new.gohtml", templateData)
}

// EditPartyHandler handles the edit party form page.
func (h *Handlers) EditPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "EditPartyHandler")
	logger.InfoContext(ctx, "calling EditPartyHandler")
	templateData := h.NewPartiesTemplateData(r, w, "/parties")
	h.renderer.Render(w, r, http.StatusOK, "parties/edit.gohtml", templateData)
}

// CreatePartyHandler handles creating a new party.
func (h *Handlers) CreatePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "CreatePartyHandler")
	profileID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get profile ID from session", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := h.NewPartiesTemplateData(r, w, "/parties")
		h.renderer.Render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "failed to get parseForm", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := h.NewPartiesTemplateData(r, w, "/parties")
		h.renderer.Render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	name := r.FormValue("partyName")
	if name == "" {
		logger.ErrorContext(ctx, "failed to get partyName from form")
		h.flashMessager.SetErrorFlashMessage(w, r, "A Name is required to create this party.")
		data := h.NewPartiesTemplateData(r, w, "/parties")
		h.renderer.Render(w, r, http.StatusBadRequest, "parties/new.gohtml", data)
		return
	}

	id, err := h.partyService.CreateParty(ctx, profileID, name)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create party", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error creating this party, try again.")
		data := h.NewPartiesTemplateData(r, w, "/parties")
		h.renderer.Render(w, r, http.StatusInternalServerError, "parties/new.gohtml", data)
		return
	}

	h.flashMessager.SetInfoFlashMessage(w, r, "Party successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/parties/%d", id), http.StatusSeeOther)
}

// PartyShowHandler handles displaying a single party with its movies and members.
func (h *Handlers) PartyShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "PartyShowHandler")

	watcher, err := h.getWatcherFromSession(ctx, r)
	if err != nil {
		h.handleFailedToGetWatcherFromSession(ctx, logger, w, r, err)
		return
	}

	idParam := r.PathValue("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get party ID from path", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusBadRequest)
		return
	}

	party, err := h.partyService.GetPartyWithMovies(ctx, logger, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get party by ID", slog.Any("error", err))
		data := h.newBaseTemplateData(r, w, "/parties")
		h.renderer.Render(w, r, http.StatusNotFound, "404.gohtml", data)
		return
	}

	currentWatcherIsOwner := watcher.ID == party.IDOwner

	invites, err := h.invitationsService.GetInvitationsForParty(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get invitations", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an issue getting this party, try again.")
		http.Redirect(w, r, "/parties", http.StatusBadRequest)
		return
	}

	templateData := h.NewPartiesTemplateData(r, w, "/parties")
	templateData.Party = party
	templateData.ModalData.PendingInvites = invites
	templateData.ModalData.PartyID = id
	templateData.CurrentWatcherIsOwner = currentWatcherIsOwner

	h.renderer.Render(w, r, http.StatusOK, "parties/show.gohtml", templateData)
}

// PartiesIndexHandler handles displaying all parties for the current user.
func (h *Handlers) PartiesIndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "PartyIndexHandler")

	watcher, err := h.getWatcherFromSession(ctx, r)
	if err != nil {
		h.handleFailedToGetWatcherFromSession(ctx, logger, w, r, err)
		return
	}

	// TODO: paginate both of these
	parties, invites, err := watcher.GetPartiesAndInvitedParties(ctx, h.partyService)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get parties and invites", slog.Any("error", err))
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an issue getting your parties, try again.")
		http.Redirect(w, r, "/profile", http.StatusBadRequest)
		return
	}

	templateData := h.NewPartiesIndexTemplateData(r, w, "/parties", parties, invites, watcher.ID)
	h.renderer.Render(w, r, http.StatusOK, "parties/index.gohtml", templateData)
}

// MarkMovieAsWatchedHandler handles marking a movie as watched in a party.
func (h *Handlers) MarkMovieAsWatchedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idPartyParam := r.PathValue("party_id")
	idMovieParam := r.PathValue("id")
	idMovie, err := strconv.Atoi(idMovieParam)
	if err != nil {
		h.renderer.ClientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		h.renderer.ClientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = h.partiesRepository.MarkPartyMovieAsWatched(ctx, idParty, idMovie)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}

// SelectMovieForParty handles selecting a random movie for a party.
func (h *Handlers) SelectMovieForParty(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "SelectMovieForParty")
	idPartyParam := r.PathValue("party_id")
	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get party ID from path", slog.Any("error", err))
		h.renderer.ClientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	err = h.partiesRepository.SelectMovieForParty(ctx, idParty)
	if err != nil {
		logger.ErrorContext(ctx, "failed to select movie for party", slog.Any("error", err))
		h.renderer.ServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/parties/"+idPartyParam, http.StatusSeeOther)
}
