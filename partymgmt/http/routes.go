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

// RegisterRoutes registers all party management routes.
// This includes movies, parties, party movies, invitations, party members, and watchers.
func (h *Handlers) RegisterRoutes() []Route {
	return []Route{
		// Movie routes
		{
			Path:               "GET /movies",
			Handler:            h.MoviesIndexHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "GET /movies/{id}",
			Handler:            h.MoviesShowHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "POST /movies/create",
			Handler:            h.MoviesCreateHandler,
			AuthenticatedRoute: false,
		},
		{
			Path:               "GET /movies/{id}/modal",
			Handler:            h.GetAddMovieToPartyModal,
			AuthenticatedRoute: true,
		},
		// Party routes
		{
			Path:               "GET /parties/",
			Handler:            h.PartiesIndexHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "GET /parties/new",
			Handler:            h.NewPartyHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "GET /parties/{id}",
			Handler:            h.PartyShowHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "POST /parties/{party_id}/movies/{id}",
			Handler:            h.MarkMovieAsWatchedHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "POST /parties/{party_id}/movies",
			Handler:            h.SelectMovieForParty,
			AuthenticatedRoute: true,
		},
		{
			Path:               "POST /parties",
			Handler:            h.CreatePartyHandler,
			AuthenticatedRoute: true,
		},
		{
			Path:               "GET /parties/{id}/edit",
			Handler:            h.EditPartyHandler,
			AuthenticatedRoute: true,
		},
		// Party movie routes
		{
			Path:               "POST /party_movies",
			Handler:            h.AddMovieToPartiesHandler,
			AuthenticatedRoute: true,
		},
		// Invitation routes
		{
			Path:               "POST /invitations",
			Handler:            h.CreateInviteHandler,
			AuthenticatedRoute: true,
		},
		// Party member routes
		{
			Path:               "POST /party_members",
			Handler:            h.AcceptInviteHandler,
			AuthenticatedRoute: true,
		},
		// Watcher routes
		{
			Path:               "GET /watched_movies",
			Handler:            h.WatchedMoviesHandler,
			AuthenticatedRoute: true,
		},
	}
}

// GetRoutes returns routes in a format compatible with the web package.
// This method exists to satisfy the web.PartyMgmtHandlers interface without creating import cycles.
func (h *Handlers) GetRoutes() []struct {
	Path               string
	Handler            http.HandlerFunc
	AuthenticatedRoute bool
} {
	routes := h.RegisterRoutes()
	result := make([]struct {
		Path               string
		Handler            http.HandlerFunc
		AuthenticatedRoute bool
	}, len(routes))

	for i, r := range routes {
		result[i] = struct {
			Path               string
			Handler            http.HandlerFunc
			AuthenticatedRoute bool
		}{
			Path:               r.Path,
			Handler:            r.Handler,
			AuthenticatedRoute: r.AuthenticatedRoute,
		}
	}

	return result
}
