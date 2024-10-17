package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/store"
)

var ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")

type MoviesService interface {
	GetMovieByTMDBID(context.Context, int) (store.Movie, error)
	GetMovieByID(context.Context, int) (store.Movie, error)
	CreateMovie(context.Context, store.Movie) (store.Movie, error)
}

type PartiesStoreService interface {
	GetPartiesByProfileForCurrentMovie(context.Context, int, int) ([]store.Party, error)
	GetPartiesForProfile(context.Context, int) ([]store.Party, error)
	GetPartyByID(context.Context, int) (store.Party, error)
	GetPartyByIDWithMovies(context.Context, int) (store.Party, error)
	AddMovieToParty(context.Context, int, int) error
	MarkMovieAsWatched(context.Context, int, int) error
	SelectMovieForParty(context.Context, int) error
}

type PartyService interface {
	CreateParty(context.Context, int, string) (int, error)
}

type ProfilesService interface {
	GetProfileByID(context.Context, int) (store.Profile, error)
}

type AccountService interface {
	CreateAccount(context.Context, string, string, string, []byte) (store.Account, error)
	FindAccountByEmail(context.Context, string) (store.Account, error)
	AccountExists(context.Context, int) (bool, error)
}

type Application struct {
	Logger              *slog.Logger
	TemplateCache       map[string]*template.Template
	TMDBClient          *TMDBClient
	SessionStore        *sessions.CookieStore
	MoviesService       MoviesService
	PartyService        PartyService
	PartiesStoreService PartiesStoreService
	ProfilesService     ProfilesService
	AccountService      AccountService
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

func (a *Application) getProfileIDFromSession(r *http.Request) (int, error) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.Logger.Error("failed to get session", slog.Any("error", err))
		return 0, nil
	}

	sessionProfileID := session.Values["profileID"]
	profileID, ok := sessionProfileID.(int)
	if !ok {
		return 0, ErrFailedToGetProfileIDFromSession
	}

	return profileID, nil
}
