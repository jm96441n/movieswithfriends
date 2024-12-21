package web

import (
	"io/fs"
	"log"
	"net/http"
	"slices"

	"github.com/jm96441n/movieswithfriends/ui"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type route struct {
	path               string
	handler            http.HandlerFunc
	authenticatedRoute bool
}

func (a *Application) Routes() http.Handler {
	router := http.NewServeMux()

	staticRoutes := a.staticRoutes()
	movieRoutes := a.movieRoutes()
	partyRoutes := a.partyRoutes()
	sessionRoutes := a.sessionRoutes()
	profileRoutes := a.profileRoutes()
	partyMemberRoutes := a.partyMemberRoutes()

	// allocate capacity for all routes
	routes := make([]route, 0, len(staticRoutes)+len(movieRoutes)+len(partyRoutes)+len(sessionRoutes)+len(profileRoutes)+len(partyMemberRoutes))

	routes = slices.Concat(
		routes,
		staticRoutes,
		movieRoutes,
		partyRoutes,
		sessionRoutes,
		profileRoutes,
		partyMemberRoutes,
	)

	authenticatorMW := a.authenticateMiddleware()
	requireAuthMW := a.authenticatedMiddleware()
	// loggingMW := loggingMiddlewareBuilder(a.Logger)

	for _, r := range routes {
		handlerFunc := r.handler
		if r.authenticatedRoute {
			handlerFunc = requireAuthMW(handlerFunc)
		}
		handlerFunc = otelhttp.NewHandler(otelhttp.WithRouteTag(r.path, authenticatorMW(handlerFunc)), r.path).(http.HandlerFunc)

		router.Handle(r.path, handlerFunc)
	}

	fsys, err := fs.Sub(ui.TemplateFS, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Create a file server handler
	fileServer := http.FileServer(http.FS(fsys))
	router.Handle("/static/", http.StripPrefix("/static", fileServer))
	return router
}

func (a *Application) staticRoutes() []route {
	return []route{
		{
			path:               "/",
			handler:            a.HomeHandler,
			authenticatedRoute: false,
		},
		{
			path: "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				return
			},
			authenticatedRoute: false,
		},
	}
}

func (a *Application) movieRoutes() []route {
	return []route{
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
	}
}

func (a *Application) partyRoutes() []route {
	return []route{
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
			path:               "PUT /parties/{id}",
			handler:            a.AddMovietoPartyHandler,
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
	}
}

func (a *Application) partyMemberRoutes() []route {
	return []route{
		{
			path:               "POST /party_members",
			handler:            a.AddMemberToPartyHandler,
			authenticatedRoute: true,
		},
	}
}

func (a *Application) sessionRoutes() []route {
	return []route{
		{
			path:               "GET /signup",
			handler:            a.SignUpShowHandler,
			authenticatedRoute: false,
		},
		{
			path:               "POST /signup",
			handler:            a.SignUpHandler,
			authenticatedRoute: false,
		},
		{
			path:               "GET /login",
			handler:            a.LoginShowHandler,
			authenticatedRoute: false,
		},
		{
			path:               "POST /login",
			handler:            a.LoginHandler,
			authenticatedRoute: false,
		},
		{
			path:               "POST /logout",
			handler:            a.LogoutHandler,
			authenticatedRoute: false,
		},
	}
}

func (a *Application) profileRoutes() []route {
	return []route{
		{
			path:               "GET /profile",
			handler:            a.ProfileShowHandler,
			authenticatedRoute: true,
		},
	}
}
