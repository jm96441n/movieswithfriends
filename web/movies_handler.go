package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
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

func (a *Application) MoviesShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idParams := vars["id"]

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
	}

	result, err := a.GetMovie(ctx, id)
	if err != nil {
		a.serverError(w, r, err)
	}

	templateData := a.NewMoviesTemplateData(r, "/movie")
	templateData.Movie = result
	a.render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}

func (a *Application) GetMovie(ctx context.Context, id int) (store.Movie, error) {
	result, err := a.MoviesService.GetMovieByTMDBID(ctx, id)
	if err == nil {
		a.Logger.Info("movie found in db", slog.Any("movie", result.Title))
		return result, nil
	}

	if !errors.Is(err, store.ErrNoRecord) {
		return store.Movie{}, fmt.Errorf("failed to retrieve movie from db: %w", err)
	}
	err = nil

	result, err = a.TMDBClient.GetMovie(ctx, id)
	if err != nil {
		return store.Movie{}, err
	}

	fmt.Println(result)

	_, err = a.MoviesService.CreateMovie(ctx, result)
	if err != nil {
		a.Logger.Error(fmt.Sprintf("Failed to create movie: %s", err), slog.Any("movie", result.Title))
		return result, nil
	}

	a.Logger.Info("movie created in db", slog.Any("movie", result))

	return result, nil
}
