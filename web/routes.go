package web

import (
	"net/http"
	"slices"
)

type route struct {
	path               string
	handler            http.HandlerFunc
	authenticatedRoute bool
}

func (a *Application) Routes() http.Handler {
	router := http.NewServeMux()

	movieRoutes := a.movieRoutes()
	partyRoutes := a.partyRoutes()
	sessionRoutes := a.sessionRoutes()
	profileRoutes := a.profileRoutes()

	routes := make([]route, 0, len(movieRoutes)+len(partyRoutes)+len(sessionRoutes)+len(profileRoutes)+1) // +1 for home route

	homeRoute := route{
		path:               "/",
		handler:            a.HomeHandler,
		authenticatedRoute: false,
	}

	routes = append(routes, homeRoute)
	routes = slices.Concat(routes, movieRoutes, partyRoutes, sessionRoutes, profileRoutes)

	authenticatorMW := a.authenticateMiddleware()
	requireAuthMW := a.authenticatedMiddleware()
	loggingMW := loggingMiddlewareBuilder(a.Logger)

	for _, r := range routes {
		handlerFunc := r.handler
		if r.authenticatedRoute {
			handlerFunc = requireAuthMW(handlerFunc)
		}
		handlerFunc = loggingMW(authenticatorMW(handlerFunc))
		router.HandleFunc(r.path, handlerFunc)
	}

	//	fileServer := http.FileServer(http.FS(ui.TemplateFS))
	//	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	return router
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
		{
			path:               "POST /profile_parties",
			handler:            a.AddFriendToPartyHandler,
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
