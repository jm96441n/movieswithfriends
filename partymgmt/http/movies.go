package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/web"
)

// NewMoviesTemplateData creates a new MoviesTemplateData with the base template data populated.
func (h *Handlers) NewMoviesTemplateData(r *http.Request, w http.ResponseWriter, path string) web.MoviesTemplateData {
	return web.MoviesTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}

// newBaseTemplateData creates the base template data by delegating to the embedded Renderer.
// This is a helper method to allow domain handlers to create template data.
func (h *Handlers) newBaseTemplateData(r *http.Request, w http.ResponseWriter, path string) web.BaseTemplateData {
	// We need to get the base template data somehow. Since Renderer doesn't expose NewTemplateData,
	// we'll need to access it through a type assertion to *web.Application.
	if app, ok := h.Renderer.(*web.Application); ok {
		return app.NewTemplateData(r, w, path)
	}
	// Fallback to empty base template data if type assertion fails
	return web.BaseTemplateData{
		CurrentPagePath: path,
		CurrentYear:     2025,
	}
}

// MoviesIndexHandler handles the movies index page with optional search functionality.
func (h *Handlers) MoviesIndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With(slog.Any("handler", "MoviesIndexHandler"))
	queryParams := r.URL.Query()
	templateData := h.NewMoviesTemplateData(r, w, "/movies")

	if _, ok := queryParams["search"]; !ok {
		h.Render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
		return
	}

	templateData.SearchValue = queryParams.Get("search")
	term := strings.TrimSpace(queryParams.Get("search"))

	movies, err := h.moviesService.SearchMovies(ctx, logger, term)
	if err != nil {
		logger.ErrorContext(ctx, "failed to search movies", slog.Any("error", err))
		h.ServerError(w, r, err)
		return
	}

	templateData.Movies = movies

	if r.Header.Get("HX-Request") != "" {
		h.RenderPartial(w, r, http.StatusOK, "movies/partials/search_results.gohtml", templateData)
		return
	}

	h.Render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
}

// MoviesCreateHandler handles creating a movie from TMDB data.
func (h *Handlers) MoviesCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With(slog.Any("handler", "MoviesCreateHandler"))
	err := r.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "failed to parse form", slog.Any("error", err))
		h.SetErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	idParams := r.FormValue("tmdb_id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		logger.ErrorContext(ctx, "failed to convert tmdb_id to int", slog.Any("error", err))
		h.SetErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
		return
	}

	movieID, err := h.moviesService.CreateMovie(ctx, logger, id)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create movie", slog.Any("error", err))
		h.SetErrorFlashMessage(w, r, "There was an error navigating to this movie, try again.")
		http.Redirect(w, r, "/movies", http.StatusSeeOther)
	}

	http.Redirect(w, r, fmt.Sprintf("/movies/%d", movieID), http.StatusSeeOther)
}

// MoviesShowHandler handles displaying a single movie's details.
func (h *Handlers) MoviesShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With(slog.Any("handler", "MoviesShowHandler"))
	idParams := r.PathValue("id")

	id, err := strconv.Atoi(idParams)
	if err != nil {
		h.ClientError(w, r, http.StatusBadRequest, "Please try again")
		return
	}

	movie, err := h.moviesService.GetMovie(ctx, logger, partymgmt.MovieID{MovieID: &id})
	if err != nil {
		if errors.Is(err, partymgmt.ErrMovieDoesNotExist) {
			logger.ErrorContext(ctx, "did not find movie in db", "id", id)
			h.SetErrorFlashMessage(w, r, "Could not find the requested movie in the database, try again")
			http.Redirect(w, r, "/movies", http.StatusNotFound)
			return
		}

		logger.ErrorContext(ctx, "failed to retrieve movie from db", "error", err)
		h.SetErrorFlashMessage(w, r, "Could not find the requested movie in the database, try again")
		http.Redirect(w, r, "/movies", http.StatusInternalServerError)
		return
	}

	templateData := h.NewMoviesTemplateData(r, w, "/movie")
	templateData.Movie = movie
	h.Render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}
