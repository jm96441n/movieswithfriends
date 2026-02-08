package http

import (
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/identityaccess/services"
)

// ProfilesTemplateDataProvider defines an interface for creating ProfilesTemplateData.
// This allows the handler to get profile-specific template data without depending on web package types.
type ProfilesTemplateDataProvider interface {
	NewProfilesTemplateData(r *http.Request, w http.ResponseWriter, path string) ProfilesTemplateData
}

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

	profileID, err := h.getProfileIDFromSession(ctx, r)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}

	page := r.URL.Query().Get("page")

	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		h.ServerError(w, r, err)
		return
	}

	movieData, err := h.profileAggregatorService.GetWatchPaginatedHistory(ctx, logger, profileID, services.PageInfo{PageNum: pageNum, PageSize: 15})
	if err != nil {
		h.ServerError(w, r, err)
		return
	}

	// Get profile-specific template data
	if provider, ok := h.Renderer.(ProfilesTemplateDataProvider); ok {
		templateData := provider.NewProfilesTemplateData(r, w, "/watched_movies")
		templateData.SetWatchedMovies(movieData.WatchedMovies)
		templateData.SetPagination(pageNum, movieData.NumPages)

		if r.Header.Get("HX-Request") != "" {
			h.RenderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
			return
		}

		h.Render(w, r, http.StatusOK, "", templateData)
		return
	}

	// This shouldn't happen in normal operation, but provides a safety net
	logger.ErrorContext(ctx, "Renderer does not implement ProfilesTemplateDataProvider interface")
	h.ServerError(w, r, ErrFailedToGetProfileIDFromSession)
}
