package web

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/jm96441n/movieswithfriends/ui"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Route struct {
	Path               string
	Handler            http.HandlerFunc
	AuthenticatedRoute bool
}

type routeRegisterer interface {
	RegisterRoutes() []Route
}

type AuthMiddlewares interface {
	AuthenticateMiddleware() func(http.HandlerFunc) http.HandlerFunc
	AuthenticatedMiddleware() func(http.HandlerFunc) http.HandlerFunc
}

func Routes(authMiddlewares AuthMiddlewares, registerers ...routeRegisterer) http.Handler {
	router := http.NewServeMux()

	authenticatorMW := authMiddlewares.AuthenticateMiddleware()
	requireAuthMW := authMiddlewares.AuthenticatedMiddleware()

	fsys, err := fs.Sub(ui.TemplateFS, "dist")
	if err != nil {
		log.Fatal(err)
	}

	// Create a file server handler
	fileServer := http.FileServer(http.FS(fsys))
	router.Handle("/dist/", http.StripPrefix("/dist", fileServer))

	for _, resisterer := range registerers {
		routes := resisterer.RegisterRoutes()
		for _, r := range routes {
			handlerFunc := r.Handler
			if r.AuthenticatedRoute {
				handlerFunc = requireAuthMW(handlerFunc)
			}
			handlerFunc = otelhttp.NewHandler(authenticatorMW(handlerFunc), r.Path).(http.HandlerFunc)

			router.Handle(r.Path, handlerFunc)
		}
	}

	return router
}
