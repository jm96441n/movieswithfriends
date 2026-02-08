package web

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	identityaccesshttp "github.com/jm96441n/movieswithfriends/identityaccess/http"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/ui"
)

var (
	ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")
	ErrFailedToGetPartyIDFromSession   = errors.New("failed to get current party id from session")
	ErrFailedToGetAccountIDFromSession = errors.New("failed to get accountl id from session")
)

// PartyMgmtHandlers defines the interface for party management HTTP handlers.
type PartyMgmtHandlers interface {
	GetRoutes() []struct {
		Path               string
		Handler            http.HandlerFunc
		AuthenticatedRoute bool
	}
}

type AppConfig struct {
	Telemetry                metrics.TelemetryProvider
	Logger                   *slog.Logger
	SessionStore             *sessions.CookieStore
	MoviesService            *partymgmt.MovieService
	MoviesRepository         *partymgmtstore.MoviesRepository
	PartyService             partymgmt.PartyService
	PartiesRepository        partymgmtstore.PartyRepository
	WatcherService           partymgmt.WatcherService
	ProfilesService          *identityaccess.ProfileService
	ProfileAggregatorService *services.ProfileAggregatorService
	InvitationsService       partymgmt.InvitationsService
	Auth                     *identityaccess.Authenticator
	AssetLoader              *Loader
	PartyMgmtHandlers        PartyMgmtHandlers
	IdentityAccessHandlers   *identityaccesshttp.Handlers
}

type Application struct {
	Telemetry                metrics.TelemetryProvider
	Logger                   *slog.Logger
	templateCache            map[string]*template.Template
	SessionStore             *sessions.CookieStore
	MoviesService            *partymgmt.MovieService
	MoviesRepository         *partymgmtstore.MoviesRepository
	PartyService             partymgmt.PartyService
	PartiesRepository        partymgmtstore.PartyRepository
	WatcherService           partymgmt.WatcherService
	ProfilesService          *identityaccess.ProfileService
	ProfileAggregatorService *services.ProfileAggregatorService
	InvitationsService       partymgmt.InvitationsService
	Auth                     *identityaccess.Authenticator
	AssetLoader              *Loader
	IdentityAccessHandlers   *identityaccesshttp.Handlers
	PartyMgmtHandlers        PartyMgmtHandlers
}

func NewApplication(cfg AppConfig) *Application {
	a := &Application{
		Telemetry:                cfg.Telemetry,
		Logger:                   cfg.Logger,
		SessionStore:             cfg.SessionStore,
		MoviesService:            cfg.MoviesService,
		MoviesRepository:         cfg.MoviesRepository,
		PartyService:             cfg.PartyService,
		PartiesRepository:        cfg.PartiesRepository,
		WatcherService:           cfg.WatcherService,
		ProfilesService:          cfg.ProfilesService,
		ProfileAggregatorService: cfg.ProfileAggregatorService,
		InvitationsService:       cfg.InvitationsService,
		Auth:                     cfg.Auth,
		AssetLoader:              cfg.AssetLoader,
		PartyMgmtHandlers:        cfg.PartyMgmtHandlers,
		IdentityAccessHandlers:   cfg.IdentityAccessHandlers,
	}

	a.initTemplateCache(ui.TemplateFS)

	return a
}

// Legacy methods - these now delegate to the public Renderer interface methods
// for backward compatibility with existing handlers in the web package.

func (a *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	a.ServerError(w, r, err)
}

func (a *Application) clientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	a.ClientError(w, r, status, message)
}

func (a *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	a.Render(w, r, status, page, data)
}

func (a *Application) renderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	a.RenderPartial(w, r, status, page, data)
}

func (a *Application) getAccountIDFromSession(ctx context.Context, r *http.Request) (int, error) {
	_, span, _ := metrics.SpanFromContext(ctx, "getAccountIDFromSession")
	defer span.End()

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		session, err = a.SessionStore.New(r, sessionName)
		if err != nil {
			a.Logger.DebugContext(ctx, "failed to create new session")
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
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.getProfileFromSession")
	defer span.End()
	profileID, err := a.getProfileIDFromSession(ctx, r)
	if err != nil {
		return nil, err
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("coudl not get profile with id: %d: %w", profileID, err)
	}

	return profile, nil
}

func (a *Application) getWatcherFromSession(ctx context.Context, r *http.Request) (partymgmt.Watcher, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Application.getWatcherFromSession")
	defer span.End()
	watcherID, err := a.getProfileIDFromSession(ctx, r)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	watcher, err := a.WatcherService.NewWatcher(ctx, watcherID)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	return watcher, nil
}

func (a *Application) getProfileIDFromSession(ctx context.Context, r *http.Request) (int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Application.getProfileIDFromSession")
	defer span.End()

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		session, err = a.SessionStore.New(r, sessionName)
		if err != nil {
			a.Logger.DebugContext(ctx, "failed to create new session")
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

func (a *Application) handleFailedToGetWatcherFromSession(ctx context.Context, logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.ErrorContext(ctx, "failed to get watcher from session", slog.Any("error", err))
	a.setErrorFlashMessage(w, r, "There was an issue authenticating you, please log in again.")
	a.logout(w, r)
	http.Redirect(w, r, "/login", http.StatusInternalServerError)
}

func (a *Application) setInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	a.SetInfoFlashMessage(w, r, msg)
}

func (a *Application) setErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	a.SetErrorFlashMessage(w, r, msg)
}

// func (a *Application) setWarningFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
// a.SetFlashMessage(w, r, FlashWarningKey, msg)
// }

func (a *Application) setFlashMessage(w http.ResponseWriter, r *http.Request, key, msg string) {
	a.SetFlashMessage(w, r, key, msg)
}

func (a *Application) logout(w http.ResponseWriter, r *http.Request) {
	_ = a.IdentityAccessHandlers.Logout(w, r)
}
