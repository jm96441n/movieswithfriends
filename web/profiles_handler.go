package web

import (
	"net/http"
)

// ProfileShowHandler delegates to the identityaccess domain handler.
func (a *Application) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.ProfileShowHandler(w, r)
}

// ProfileEditPageHandler delegates to the identityaccess domain handler.
func (a *Application) ProfileEditPageHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.ProfileEditPageHandler(w, r)
}

// ProfileEditHandler delegates to the identityaccess domain handler.
func (a *Application) ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.ProfileEditHandler(w, r)
}

// GetPaginatedWatchHistoryHandler delegates to the identityaccess domain handler.
func (a *Application) GetPaginatedWatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.GetPaginatedWatchHistoryHandler(w, r)
}
