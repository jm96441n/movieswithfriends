package web

import (
	"io/fs"
	"log"
	"net/http"
	"slices"

	"github.com/jm96441n/movieswithfriends/ui"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Route struct {
	path               string
	handler            http.HandlerFunc
	authenticatedRoute bool
}

func (a *Application) Routes() http.Handler {
	router := http.NewServeMux()

	staticRoutes := a.staticRoutes()
	movieRoutes := a.movieRoutes()
	partyRoutes := a.partyRoutes()
	partyMovieRoutes := a.partyMovieRoutes()
	partyMemberRoutes := a.partyMemberRoutes()
	invitationRoutes := a.invitationRoutes()

	// Domain-specific routes
	identityAccessRoutes := a.identityAccessRoutes()

	// allocate capacity for all routes
	routes := make([]Route, 0)

	routes = slices.Concat(
		routes,
		staticRoutes,
		movieRoutes,
		partyRoutes,
		partyMovieRoutes,
		invitationRoutes,
		partyMemberRoutes,
		identityAccessRoutes,
	)

	authenticatorMW := a.authenticateMiddleware()
	requireAuthMW := a.authenticatedMiddleware()

	fsys, err := fs.Sub(ui.TemplateFS, "dist")
	if err != nil {
		log.Fatal(err)
	}

	// Create a file server handler
	fileServer := http.FileServer(http.FS(fsys))
	router.Handle("/dist/", http.StripPrefix("/dist", fileServer))

	for _, r := range routes {
		handlerFunc := r.handler
		if r.authenticatedRoute {
			handlerFunc = requireAuthMW(handlerFunc)
		}
		handlerFunc = otelhttp.NewHandler(otelhttp.WithRouteTag(r.path, authenticatorMW(handlerFunc)), r.path).(http.HandlerFunc)

		router.Handle(r.path, handlerFunc)
	}

	return router
}

func (a *Application) staticRoutes() []Route {
	return []Route{
		{
			path:               "/",
			handler:            a.HomeHandler,
			authenticatedRoute: false,
		},
		{
			path: "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			authenticatedRoute: false,
		},
	}
}

func (a *Application) movieRoutes() []Route {
	return []Route{
		{
			path:               "GET /movies",
			handler:            a.MoviesIndexHandler,
			authenticatedRoute: false,
		},
		{
			path:               "GET /movies/{id}",
			handler:            a.MoviesShowHandler,
			authenticatedRoute: false,
		},
		{
			path:               "POST /movies/create",
			handler:            a.MoviesCreateHandler,
			authenticatedRoute: false,
		},
		{
			path:               "GET /movies/{id}/modal",
			handler:            a.PartyMgmtHandlers.GetAddMovieToPartyModal,
			authenticatedRoute: true,
		},
	}
}

func (a *Application) partyMovieRoutes() []Route {
	return []Route{
		{
			path:               "POST /party_movies",
			handler:            a.PartyMgmtHandlers.AddMovieToPartiesHandler,
			authenticatedRoute: true,
		},
	}
}

func (a *Application) partyRoutes() []Route {
	return []Route{
		{
			path:               "GET /parties/",
			handler:            a.PartiesIndexHandler,
			authenticatedRoute: true,
		},
		{
			path:               "GET /parties/new",
			handler:            a.NewPartyHandler,
			authenticatedRoute: true,
		},
		{
			path:               "GET /parties/{id}",
			handler:            a.PartyShowHandler,
			authenticatedRoute: true,
		},
		{
			path:               "POST /parties/{party_id}/movies/{id}",
			handler:            a.MarkMovieAsWatchedHandler,
			authenticatedRoute: true,
		},
		{
			path:               "POST /parties/{party_id}/movies",
			handler:            a.SelectMovieForParty,
			authenticatedRoute: true,
		},
		{
			path:               "POST /parties",
			handler:            a.CreatePartyHandler,
			authenticatedRoute: true,
		},
		{
			path:               "GET /parties/{id}/edit",
			handler:            a.EditPartyHandler,
			authenticatedRoute: true,
		},
	}
}

func (a *Application) invitationRoutes() []Route {
	return []Route{
		{
			path:               "POST /invitations",
			handler:            a.CreateInviteHandler,
			authenticatedRoute: true,
		},
	}
}

func (a *Application) identityAccessRoutes() []Route {
	// Get routes from the identityaccess domain package
	domainRoutes := a.IdentityAccessHandlers.RegisterRoutes()

	// Convert domain routes to web routes
	routes := make([]Route, len(domainRoutes))
	for i, dr := range domainRoutes {
		routes[i] = Route{
			path:               dr.Path,
			handler:            dr.Handler,
			authenticatedRoute: dr.AuthenticatedRoute,
		}
	}

	return routes
}

func (a *Application) partyMemberRoutes() []Route {
	return []Route{
		{
			path:               "POST /party_members",
			handler:            a.AcceptInviteHandler,
			authenticatedRoute: true,
		},
	}
}

// func (a *Application) watcherRoutes() []Route {
// 	return []Route{
// 		{
// 			path:               "GET /watched_movies",
// 			handler:            a.WatchedMoviesHandler,
// 			authenticatedRoute: true,
// 		},
// 	}
// }
