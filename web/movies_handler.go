package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) MoviesIndexHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With(slog.Any("handler", "MoviesIndexHandler"))
	queryParams := r.URL.Query()
	templateData := a.NewMoviesTemplateData(r, w, "/movies")

	if _, ok := queryParams["search"]; !ok {
		a.render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
		return
	}

	ctx := r.Context()
	templateData.SearchValue = queryParams.Get("search")
	term := strings.TrimSpace(queryParams.Get("search"))

	movies, err := a.MoviesService.SearchMovies(ctx, logger, term)
	if err != nil {
		logger.ErrorContext(ctx, "failed to search movies", slog.Any("error", err))
		a.serverError(w, r, err)
	}

	templateData.Movies = movies
	if r.Header.Get("HX-Request") != "" {
		a.renderPartial(w, r, http.StatusOK, "movies/partials/search_results.gohtml", templateData)
		return
	}

	a.render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
}

func (a *Application) MoviesCreateHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With(slog.Any("handler", "MoviesCreateHandler"))
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "failed to parse form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	idParams := r.FormValue("tmdb_id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		logger.ErrorContext(ctx, "failed to convert tmdb_id to int", slog.Any("error", err))
		a.clientError(w, http.StatusBadRequest)
		return
	}

	movie, err := a.MoviesService.CreateMovie(ctx, logger, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create movie", slog.Any("error", err))
		a.serverError(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/movies/%d", movie.ID), http.StatusSeeOther)
}

func (a *Application) MoviesShowHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With(slog.Any("handler", "MoviesShowHandler"))
	ctx := r.Context()
	idParams := r.PathValue("id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	memberID, err := a.getProfileIDFromSession(r)
	if errors.Is(err, ErrFailedToGetProfileIDFromSession) {
		logger.DebugContext(ctx, "profileID is not in session")
	} else if err != nil {
		logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	result, err := a.MoviesRepository.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			logger.ErrorContext(ctx, "did not find movie in db", "id", id)
			a.clientError(w, http.StatusNotFound)
			return
		}

		logger.ErrorContext(ctx, "failed to retrieve movie from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	parties, err := a.PartiesRepository.GetPartiesByMemberIDForCurrentMovie(ctx, id, memberID)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get parties", "error", err)
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewMoviesTemplateData(r, w, "/movie")
	templateData.Movie = result
	templateData.Parties = parties
	a.render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}
