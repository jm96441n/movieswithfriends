package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

func (a *Application) MoviesIndexHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r)
	a.render(w, r, http.StatusOK, "movies/index.gohtml", data)
}

func (a *Application) MoviesSearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
	}

	term, err := formatTerm(body)
	if err != nil {
		a.Logger.Error(err.Error())
		a.clientError(w, http.StatusBadRequest)
	}

	result, err := a.TMDBClient.Search(ctx, term, 1)
	if err != nil {
		a.serverError(w, r, err)
	}

	for idx := range result.Movies {
		result.Movies[idx].URL = fmt.Sprintf("/movies/%d", result.Movies[idx].TMDBID)
		result.Movies[idx].PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s", result.Movies[idx].PosterURL)
	}

	templateData := a.NewMoviesTemplateData(r)
	templateData.Movies = result.Movies
	a.renderPartial(w, r, http.StatusOK, "movies/partials/search_results.gohtml", templateData)
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

	templateData := a.NewMoviesTemplateData(r)
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

func formatTerm(body []byte) (string, error) {
	term := string(body)
	term, found := strings.CutPrefix(term, "search=")

	if !found {
		return "", fmt.Errorf("term not found")
	}

	return term, nil
}
