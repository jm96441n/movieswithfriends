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

	// Domain-specific routes
	identityAccessRoutes := a.identityAccessRoutes()
	partyMgmtRoutes := a.partyMgmtRoutes()

	// allocate capacity for all routes
	routes := make([]Route, 0)

	routes = slices.Concat(
		routes,
		staticRoutes,
		identityAccessRoutes,
		partyMgmtRoutes,
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

func (a *Application) partyMgmtRoutes() []Route {
	// Get routes from the partymgmt domain package
	domainRoutes := a.PartyMgmtHandlers.GetRoutes()

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
