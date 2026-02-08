package http

import (
	"net/http"
)

// Route represents an HTTP route with its configuration.
type Route struct {
	Path               string
	Handler            http.HandlerFunc
	AuthenticatedRoute bool
}

// RegisterRoutes registers all identity and access management routes.
// This includes authentication (login/logout/signup) and profile management routes.
func (h *Handlers) RegisterRoutes() []Route {
	return []Route{
		// Authentication routes
		{
			Path:               "GET /signup",
			Handler:            h.SignUpShowHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "POST /signup",
			Handler:            h.SignUpHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "GET /login",
			Handler:            h.LoginShowHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "POST /login",
			Handler:            h.LoginHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "POST /logout",
			Handler:            h.LogoutHandler,
			AuthenticatedRoute: false,
		},
		// Profile management routes
		{
			Path:               "GET /profile",
			Handler:            h.ProfileShowHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "GET /profile/edit",
			Handler:            h.ProfileEditPageHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "POST /profile",
			Handler:            h.ProfileEditHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "GET /profile/watched",
			Handler:            h.GetPaginatedWatchHistoryHandler,
			AuthenticatedRoute: true,
		},
	}
}
