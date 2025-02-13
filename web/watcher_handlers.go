package web

import (
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/identityaccess/services"
)

func (a *Application) WatchedMoviesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.GetLogger(ctx).With("handler", "WatchedMoviesHandler")
	logger.DebugContext(ctx, "getting paginated movies list")

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	page := r.URL.Query().Get("page")

	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	movieData, err := a.ProfileAggregatorService.GetWatchPaginatedHistory(ctx, logger, profileID, services.PageInfo{PageNum: pageNum, PageSize: 15})
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewProfilesTemplateData(r, w, "/watched_movies")
	templateData.WatchedMovies = movieData.WatchedMovies
	templateData.CurPage = pageNum
	templateData.NumPages = movieData.NumPages

	if r.Header.Get("HX-Request") != "" {
		a.renderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
		return
	}

	a.render(w, r, http.StatusOK, "", templateData)
}
