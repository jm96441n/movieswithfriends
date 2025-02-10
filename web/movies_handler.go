package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jm96441n/movieswithfriends/partymgmt"
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
		return
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
		a.setErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	idParams := r.FormValue("tmdb_id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		logger.ErrorContext(ctx, "failed to convert tmdb_id to int", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	movieID, err := a.MoviesService.CreateMovie(ctx, logger, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create movie", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
	}

	http.Redirect(w, r, fmt.Sprintf("/movies/%d", movieID), http.StatusSeeOther)
}

func (a *Application) MoviesShowHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With(slog.Any("handler", "MoviesShowHandler"))
	ctx := r.Context()
	idParams := r.PathValue("id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "Please try again")
		return
	}

	movie, err := a.MoviesService.GetMovie(ctx, logger, partymgmt.MovieID{MovieID: &id})
	if err != nil {
		if errors.Is(err, partymgmt.ErrMovieDoesNotExist) {
			logger.ErrorContext(ctx, "did not find movie in db", "id", id)
			a.setErrorFlashMessage(w, r, "Could not find the requested movie in the database, try again")
			http.Redirect(w, r, "/movies", http.StatusNotFound)
			return
		}

		logger.ErrorContext(ctx, "failed to retrieve movie from db", "error", err)
		a.setErrorFlashMessage(w, r, "Could not find the requested movie in the database, try again")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
		return
	}

	templateData := a.NewMoviesTemplateData(r, w, "/movie")
	templateData.Movie = movie
	a.render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}
