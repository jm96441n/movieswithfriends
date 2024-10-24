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
	queryParams := r.URL.Query()
	templateData := a.NewMoviesTemplateData(r, "/movies")

	if _, ok := queryParams["search"]; !ok {
		a.render(w, r, http.StatusOK, "movies/index.gohtml", templateData)
		return
	}

	ctx := r.Context()
	templateData.SearchValue = queryParams.Get("search")
	term := strings.TrimSpace(queryParams.Get("search"))

	movies, err := a.MoviesService.SearchMovies(ctx, term)
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

	movie, err := a.MoviesService.CreateMovie(ctx, id)
	if err != nil {
		a.serverError(w, r, err)
	}

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

	memberID, err := a.getProfileIDFromSession(r)
	if errors.Is(err, ErrFailedToGetProfileIDFromSession) {
		a.Logger.Debug("profileID is not in session")
	} else if err != nil {
		a.Logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	result, err := a.MoviesRepository.GetMovieByID(ctx, id)
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

	parties, err := a.PartiesRepository.GetPartiesByMemberIDForCurrentMovie(ctx, id, memberID)
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
