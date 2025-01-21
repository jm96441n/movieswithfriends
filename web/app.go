package web

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/ui"
)

var (
	ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")
	ErrFailedToGetPartyIDFromSession   = errors.New("failed to get current party id from session")
	ErrFailedToGetAccountIDFromSession = errors.New("failed to get accountl id from session")
)

type AppConfig struct {
	Logger                   *slog.Logger
	SessionStore             *sessions.CookieStore
	MoviesService            *partymgmt.MovieService
	MoviesRepository         *partymgmtstore.MoviesRepository
	PartyService             *partymgmt.PartyService
	PartiesRepository        *partymgmtstore.PartyRepository
	WatcherService           *partymgmt.WatcherService
	ProfilesService          *identityaccess.ProfileService
	ProfileAggregatorService *services.ProfileAggregatorService
	Auth                     *identityaccess.Authenticator
	AssetLoader              *Loader
}

type Application struct {
	Logger                   *slog.Logger
	templateCache            map[string]*template.Template
	SessionStore             *sessions.CookieStore
	MoviesService            *partymgmt.MovieService
	MoviesRepository         *partymgmtstore.MoviesRepository
	PartyService             *partymgmt.PartyService
	PartiesRepository        *partymgmtstore.PartyRepository
	WatcherService           *partymgmt.WatcherService
	ProfilesService          *identityaccess.ProfileService
	ProfileAggregatorService *services.ProfileAggregatorService
	Auth                     *identityaccess.Authenticator
	AssetLoader              *Loader
}

func NewApplication(cfg AppConfig) *Application {
	a := &Application{
		Logger:                   cfg.Logger,
		SessionStore:             cfg.SessionStore,
		MoviesService:            cfg.MoviesService,
		MoviesRepository:         cfg.MoviesRepository,
		PartyService:             cfg.PartyService,
		PartiesRepository:        cfg.PartiesRepository,
		WatcherService:           cfg.WatcherService,
		ProfilesService:          cfg.ProfilesService,
		ProfileAggregatorService: cfg.ProfileAggregatorService,
		Auth:                     cfg.Auth,
		AssetLoader:              cfg.AssetLoader,
	}

	a.initTemplateCache(ui.TemplateFS)

	return a
}

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	a.Logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (a *Application) clientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.serverError(w, r, err)
	}
	session.AddFlash(message)
	session.Save(r, w)

	http.Error(w, http.StatusText(status), status)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.templateCache[page]

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

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (a *Application) renderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.templateCache[page]

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

func (a *Application) getAccountIDFromSession(r *http.Request) (int, error) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		session, err = a.SessionStore.New(r, sessionName)
		if err != nil {
			a.Logger.Debug("failed to create new session")
			return 0, nil
		}
	}

	sessionAccountID := session.Values["accountID"]
	accountID, ok := sessionAccountID.(int)
	if !ok {
		return 0, ErrFailedToGetAccountIDFromSession
	}

	return accountID, nil
}

func (a *Application) getProfileFromSession(r *http.Request) (*identityaccess.Profile, error) {
	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		return nil, err
	}

	profile, err := a.ProfilesService.GetProfileByID(r.Context(), profileID)
	if err != nil {
		return nil, fmt.Errorf("coudl not get profile with id: %d: %w", profileID, err)
	}

	return profile, nil
}

func (a *Application) getWatcherFromSession(r *http.Request) (partymgmt.Watcher, error) {
	watcherID, err := a.getProfileIDFromSession(r)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	watcher, err := a.WatcherService.GetWatcher(r.Context(), watcherID)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	return watcher, nil
}

func (a *Application) getProfileIDFromSession(r *http.Request) (int, error) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		session, err = a.SessionStore.New(r, sessionName)
		if err != nil {
			a.Logger.Debug("failed to create new session")
			return 0, nil
		}
	}

	sessionProfileID := session.Values["profileID"]
	profileID, ok := sessionProfileID.(int)
	if !ok {
		return 0, ErrFailedToGetProfileIDFromSession
	}

	return profileID, nil
}

func (a *Application) getCurrentPartyFromSession(r *http.Request) (partymgmt.Party, error) {
	currentPartyID, err := a.getCurrentPartyIDFromSession(r)
	if err != nil {
		return partymgmt.Party{}, err
	}

	res, err := a.PartiesRepository.GetPartyByID(r.Context(), currentPartyID)
	if err != nil {
		return partymgmt.Party{}, err
	}

	party := partymgmt.Party{
		ID:      currentPartyID,
		Name:    res.Name,
		ShortID: res.ShortID,
		DB:      a.PartiesRepository,
	}

	return party, nil
}

func (a *Application) getCurrentPartyIDFromSession(r *http.Request) (int, error) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		session, err = a.SessionStore.New(r, sessionName)
		if err != nil {
			a.Logger.Debug("failed to create new session")
			return 0, nil
		}
	}

	sessionPartyID := session.Values["currentPartyID"]
	partyID, ok := sessionPartyID.(int)
	if !ok {
		return 0, ErrFailedToGetPartyIDFromSession
	}

	return partyID, nil
}

func (a *Application) setInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	a.setFlashMessage(w, r, FlashInfoKey, msg)
}

func (a *Application) setErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	a.setFlashMessage(w, r, FlashErrorKey, msg)
}

func (a *Application) setWarningFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	a.setFlashMessage(w, r, FlashWarningKey, msg)
}

func (a *Application) setFlashMessage(w http.ResponseWriter, r *http.Request, key, msg string) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.Logger.Error("failed to get session", slog.Any("error", err))
		return
	}

	a.Logger.Debug("adding flash message")
	session.AddFlash(msg, key)
	session.Save(r, w)
}
