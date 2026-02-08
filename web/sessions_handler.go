package web

import (
	"net/http"
)

// LoginShowHandler delegates to the identityaccess domain handler.
func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.LoginShowHandler(w, r)
}

// LoginHandler delegates to the identityaccess domain handler.
func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.LoginHandler(w, r)
}

// LogoutHandler delegates to the identityaccess domain handler.
func (a *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	a.IdentityAccessHandlers.LogoutHandler(w, r)
}

// logout is a helper method that calls the domain handler's logout functionality.
// This is used internally by other handlers (middleware, profile handlers, etc.)
// when they need to log out the user.
func (a *Application) logout(w http.ResponseWriter, r *http.Request) error {
	return a.IdentityAccessHandlers.Logout(w, r)
}
