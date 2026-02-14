package http

import (
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/identityaccess/services"
)

// ProfilesTemplateData represents template data for profile pages.
// This interface matches the structure used for rendering watch history.
type ProfilesTemplateData interface {
	SetWatchedMovies(movies interface{})
	SetPagination(curPage, numPages int)
}

// WatchedMoviesHandler displays paginated watch history for the authenticated user.
func (h *Handlers) WatchedMoviesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "WatchedMoviesHandler")
	logger.DebugContext(ctx, "getting paginated watch history")

	profileID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	page := r.URL.Query().Get("page")

	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	movieData, err := h.profileAggregatorService.GetWatchPaginatedHistory(ctx, logger, profileID, services.PageInfo{PageNum: pageNum, PageSize: 15})
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	// Get profile-specific template data
	templateData := NewWatchListTemplateData(r, w, "/watched_movies", h.renderer)
	templateData.WatchedMovies = movieData.WatchedMovies
	templateData.NumPages = movieData.NumPages
	templateData.CurPage = pageNum

	if r.Header.Get("HX-Request") != "" {
		h.renderer.RenderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
		return
	}

	// This shouldn't happen in normal operation, but provides a safety net
	h.renderer.Render(w, r, http.StatusOK, "", templateData)
}
