package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/partymgmt"
)

func (a *Application) AddMovieToPartiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "AddMovieToPartiesHandler")
	err := r.ParseForm()
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	watcher, err := a.getWatcherFromSession(ctx, r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get watcher from session", slog.Any("error", err))
		a.clientError(w, r, http.StatusInternalServerError, "failed to get watcher from session")
		return
	}

	formMovieID := r.FormValue("movie_id")
	formTMDBID := r.FormValue("tmdb_id")
	if formMovieID == "" && formTMDBID == "" {
		logger.ErrorContext(ctx, "no movie id or tmdb id")
		a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
		return
	}

	partyIDs := r.Form["party_ids[]"]

	logger.InfoContext(ctx, "party ids", slog.Any("party_ids", partyIDs))

	var mgmtMovieID partymgmt.MovieID

	if formMovieID != "" {
		id, err := strconv.Atoi(formMovieID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to convert movie id to int", slog.Any("error", err))
			a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
			return
		}
		mgmtMovieID.MovieID = &id
	}

	if formTMDBID != "" {
		id, err := strconv.Atoi(formTMDBID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to convert tmdb id to int", slog.Any("error", err))
			a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
			return
		}
		mgmtMovieID.TMDBID = &id
	}

	movieID, err := a.MoviesService.GetOrCreateMovie(ctx, logger, mgmtMovieID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get or create movie", slog.Any("error", err))
		a.clientError(w, r, http.StatusBadRequest, "error creating movie")
		return
	}

	for _, partyID := range partyIDs {
		id, err := strconv.Atoi(partyID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to convert party id to int", slog.Any("error", err))
			a.clientError(w, r, http.StatusBadRequest, "error creating movie")
			return
		}

		party := a.PartyService.NewParty(ctx, id, "", 0, 0, 0)
		party.ID = id
		party.AddMovie(ctx, watcher.ID, movieID)
	}

	w.Header().Set("HX-Trigger", "MovieAddedToParties")

	logger.InfoContext(ctx, "successfully added movie to parties")
}

func (a *Application) GetAddMovieToPartyModal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "GetAddMovieToMovieModal")
	idParams := r.PathValue("id")

	movieID, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "Please try again")
		return
	}

	queryParams := r.URL.Query()
	typeParam := queryParams.Get("type")

	var mgmtID partymgmt.MovieID
	var tmplData AddMovieToPartiesModalTemplateData
	switch typeParam {
	case "tmdb":
		mgmtID.TMDBID = &movieID
		tmplData.TMDBID = movieID
	case "id":
		mgmtID.MovieID = &movieID
		tmplData.MovieID = movieID
	default:
	}

	watcher, err := a.getWatcherFromSession(ctx, r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get watcher from session", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
		return
	}

	parties, err := watcher.GetPartiesToAddMovie(ctx, logger, mgmtID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get parties to add movie", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
		return
	}

	tmplData.AddedParties = parties.WithMovie
	tmplData.NotAddedParties = parties.WithoutMovie

	a.renderPartial(w, r, http.StatusOK, "movies/partials/add_to_party_modal.gohtml", tmplData)
}
