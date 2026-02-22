package web

import (
	"net/http"
)

type StaticRoutes struct {
	Renderer            *Renderer
	NewBaseTemplateData func(*http.Request, http.ResponseWriter, string) BaseTemplateData
}

func (s *StaticRoutes) RegisterRoutes() []Route {
	return []Route{
		{
			Path:               "/",
			Handler:            s.HomeHandler,
			AuthenticatedRoute: false,
		},
		{
			Path: "/health",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			AuthenticatedRoute: false,
		},
	}
}

func (s *StaticRoutes) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		data := s.NewBaseTemplateData(r, w, "/")
		s.Renderer.Render(w, r, http.StatusNotFound, "404.gohtml", data)
		return
	}
	data := s.NewBaseTemplateData(r, w, "/")
	s.Renderer.Render(w, r, http.StatusOK, "home.gohtml", data)
}
