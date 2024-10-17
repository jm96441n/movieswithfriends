package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) MoviesIndexHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	templateData := a.NewMoviesTemplateData(r, "/movies")

	if _, ok := queryParams["search"]; !ok {
		a.render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
		return
	}

	ctx := r.Context()
	templateData.SearchValue = queryParams.Get("search")

	movies, err := a.searchMovies(ctx, queryParams)
	if err != nil {
		a.serverError(w, r, err)
	}

	templateData.Movies = movies
	if r.Header.Get("HX-Request") != "" {
		a.renderPartial(w, r, http.StatusOK, "movies/partials/search_results.gohtml", templateData)
		return
	}

	a.render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
}

func (a *Application) searchMovies(ctx context.Context, queryParams url.Values) ([]store.Movie, error) {
	// handle searching for movies
	term := strings.TrimSpace(queryParams.Get("search"))

	result, err := a.TMDBClient.Search(ctx, term, 1)
	if err != nil {
		return nil, err
	}

	for idx := range result.Movies {
		result.Movies[idx].URL = fmt.Sprintf("/movies/%d", result.Movies[idx].TMDBID)
		result.Movies[idx].PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s", result.Movies[idx].PosterURL)
	}

	return result.Movies, nil
}

func (a *Application) MoviesCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	idParams := r.FormValue("tmdb_id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	movie, err := a.MoviesService.GetMovieByTMDBID(ctx, id)
	if err == nil {
		a.Logger.Info("movie found in db", slog.Any("movie", movie.Title))
		http.Redirect(w, r, fmt.Sprintf("/movies/%d", movie.ID), http.StatusSeeOther)
		return
	}

	if !errors.Is(err, store.ErrNoRecord) {
		a.serverError(w, r, err)
		return
	}
	err = nil

	movie, err = a.TMDBClient.GetMovie(ctx, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	movie, err = a.MoviesService.CreateMovie(ctx, movie)
	if err != nil {
		a.Logger.Error(fmt.Sprintf("Failed to create movie: %s", err), slog.Any("movie", movie.Title))
		a.serverError(w, r, err)
		return
	}

	a.Logger.Info("movie created in db", slog.Any("movie", movie))
	http.Redirect(w, r, fmt.Sprintf("/movies/%d", movie.ID), http.StatusSeeOther)
}

func (a *Application) MoviesShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParams := r.PathValue("id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	profileID, err := a.getProfileIDFromSession(r)
	if errors.Is(err, ErrFailedToGetProfileIDFromSession) {
		a.Logger.Debug("profileID is not in session")
	} else if err != nil {
		a.Logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	result, err := a.MoviesService.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find movie in db", "id", id)
			a.clientError(w, http.StatusNotFound)
			return
		}

		a.Logger.Error("failed to retrieve movie from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	parties, err := a.PartiesStoreService.GetPartiesByProfileForCurrentMovie(ctx, id, profileID)
	if err != nil {
		a.Logger.Error("failed to get parties", "error", err)
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewMoviesTemplateData(r, "/movie")
	templateData.Movie = result
	templateData.Parties = parties
	a.render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}
