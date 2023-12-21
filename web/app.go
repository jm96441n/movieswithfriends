package web

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

type MoviesService interface {
	GetMovieByTMDBID(context.Context, int) (store.Movie, error)
	GetMovieByID(context.Context, int) (store.Movie, error)
	CreateMovie(context.Context, store.Movie) (store.Movie, error)
}

type PartiesService interface {
	GetParties(context.Context) ([]store.Party, error)
}

type Application struct {
	Logger         *slog.Logger
	TemplateCache  map[string]*template.Template
	TMDBClient     *TMDBClient
	MoviesService  MoviesService
	PartiesService PartiesService
}

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	a.Logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.TemplateCache[page]

	if !ok {
		a.serverError(w, r, fmt.Errorf("template does not exist for page %q", page))
	}

	buf := bytes.NewBuffer([]byte{})

	// write to the buffer  to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Info("rendering template", slog.Any("template", page))
	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (a *Application) renderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.TemplateCache[page]

	if !ok {
		a.Logger.Error("template does not exist for page", slog.Any("page", page))
		a.serverError(w, r, fmt.Errorf("template does not exist for page %q", page))
		return
	}

	tmplName := filepath.Base(page)

	buf := bytes.NewBuffer([]byte{})
	// write to the buffer  to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Info("rendering template", slog.Any("template", page))

	w.WriteHeader(status)

	buf.WriteTo(w)
}
