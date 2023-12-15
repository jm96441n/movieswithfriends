package web

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func (a *Application) MoviesIndexHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r)
	a.render(w, r, http.StatusOK, "movies/index.gohtml", data)
}

func (a *Application) MoviesSearchHandler(w http.ResponseWriter, r *http.Request) {
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

	result, err := a.TMDBClient.Search(term, 1)
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
	vars := mux.Vars(r)
	idParams := vars["id"]

	id, err := strconv.Atoi(idParams)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
	}

	result, err := a.TMDBClient.GetMovie(id)
	if err != nil {
		a.serverError(w, r, err)
	}

	templateData := a.NewMoviesTemplateData(r)
	templateData.Movie = result
	a.render(w, r, http.StatusOK, "movies/show.gohtml", templateData)
}

func formatTerm(body []byte) (string, error) {
	term := string(body)
	term, found := strings.CutPrefix(term, "search=")

	if !found {
		return "", fmt.Errorf("term not found")
	}

	return term, nil
}
