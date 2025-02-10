package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/partymgmt"
)

func (a *Application) AddMovietoPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "AddMovieToPartyHandler")
	err := r.ParseForm()
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	formMovieID := r.FormValue("id_movie")
	formTMDBID := r.FormValue("tmdb_id")
	if formMovieID == "" && formTMDBID == "" {
		logger.Error("no movie id or tmdb id")
		a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
		return
	}

	var mgmtMovieID partymgmt.MovieID

	if formMovieID != "" {
		id, err := strconv.Atoi(formMovieID)
		if err != nil {
			logger.Error("failed to convert movie id to int", slog.Any("error", err))
			a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
			return
		}
		mgmtMovieID.MovieID = &id
	}

	if formTMDBID != "" {
		id, err := strconv.Atoi(formTMDBID)
		if err != nil {
			logger.Error("failed to convert tmdb id to int", slog.Any("error", err))
			a.clientError(w, r, http.StatusBadRequest, "no movie id or tmdb id")
			return
		}
		mgmtMovieID.TMDBID = &id
	}

	_, err = a.MoviesService.GetOrCreateMovie(ctx, logger, mgmtMovieID)
	if err != nil {
		logger.Error("failed to get or create movie", slog.Any("error", err))
		a.clientError(w, r, http.StatusBadRequest, "error creating movie")
		return
	}

	// err = currentParty.AddMovie(ctx, watcher.ID, id)
	// if err != nil {
	// logger.Error("failed to add movie to party", slog.Any("error", err))
	// a.clientError(w, r, http.StatusBadRequest, "error creating movie")
	// return
	// }

	partial := "movies/partials/added_movie_button_search.gohtml"
	if formMovieID != "" {
		partial = "movies/partials/added_movie_button_show.gohtml"
	}

	logger.Info("successfully added movie to party")
	a.renderPartial(w, r, http.StatusOK, partial, nil)
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

	watcher, err := a.getWatcherFromSession(r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get watcher from session", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
		return
	}

	parties, err := watcher.GetPartiesToAddMovie(ctx, logger, movieID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get parties to add movie", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an issue getting this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
	}

	tmplData := AddMovieToPartiesModalTemplateData{
		AddedParties:    parties.WithMovie,
		NotAddedParties: parties.WithoutMovie,
	}

	a.renderPartial(w, r, http.StatusOK, "movies/partials/add_to_party_modal.gohtml", tmplData)
}
