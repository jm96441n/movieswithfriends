package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/web"
)

var (
	ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")
)

// Handlers holds all HTTP handlers for the partymgmt domain.
// It embeds web.Renderer for template rendering and flash messages,
// and contains domain-specific services for business logic.
type Handlers struct {
	web.Renderer
	logger                   *slog.Logger
	sessionStore             *sessions.CookieStore
	moviesService            *partymgmt.MovieService
	partyService             partymgmt.PartyService
	watcherService           partymgmt.WatcherService
	invitationsService       partymgmt.InvitationsService
	partiesRepository        store.PartyRepository
	profileAggregatorService *services.ProfileAggregatorService
	telemetry                metrics.TelemetryProvider
}

// Config contains the dependencies required to create Handlers.
type Config struct {
	Renderer                 web.Renderer
	Logger                   *slog.Logger
	SessionStore             *sessions.CookieStore
	MoviesService            *partymgmt.MovieService
	PartyService             partymgmt.PartyService
	WatcherService           partymgmt.WatcherService
	InvitationsService       partymgmt.InvitationsService
	PartiesRepository        store.PartyRepository
	ProfileAggregatorService *services.ProfileAggregatorService
	Telemetry                metrics.TelemetryProvider
}

// NewHandlers creates a new Handlers instance with the provided dependencies.
func NewHandlers(cfg Config) *Handlers {
	return &Handlers{
		Renderer:                 cfg.Renderer,
		logger:                   cfg.Logger,
		sessionStore:             cfg.SessionStore,
		moviesService:            cfg.MoviesService,
		partyService:             cfg.PartyService,
		watcherService:           cfg.WatcherService,
		invitationsService:       cfg.InvitationsService,
		partiesRepository:        cfg.PartiesRepository,
		profileAggregatorService: cfg.ProfileAggregatorService,
		telemetry:                cfg.Telemetry,
	}
}

// Session helper methods

const sessionName = "movieswithfriends"

// getProfileIDFromSession retrieves the profile ID from the session.
func (h *Handlers) getProfileIDFromSession(ctx context.Context, r *http.Request) (int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Handlers.getProfileIDFromSession")
	defer span.End()

	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		session, err = h.sessionStore.New(r, sessionName)
		if err != nil {
			h.logger.DebugContext(ctx, "failed to create new session")
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

// getWatcherFromSession retrieves the watcher from the session.
func (h *Handlers) getWatcherFromSession(ctx context.Context, r *http.Request) (partymgmt.Watcher, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Handlers.getWatcherFromSession")
	defer span.End()

	watcherID, err := h.getProfileIDFromSession(ctx, r)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	watcher, err := h.watcherService.NewWatcher(ctx, watcherID)
	if err != nil {
		return partymgmt.Watcher{}, err
	}

	return watcher, nil
}

// handleFailedToGetWatcherFromSession is a helper for handling authentication failures.
func (h *Handlers) handleFailedToGetWatcherFromSession(ctx context.Context, logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	logger.ErrorContext(ctx, "failed to get watcher from session", slog.Any("error", err))
	h.SetErrorFlashMessage(w, r, "There was an issue authenticating you, please log in again.")
	h.logout(w, r)
	http.Redirect(w, r, "/login", http.StatusInternalServerError)
}

// logout clears the session.
func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Handlers.logout")
	defer span.End()

	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get session", slog.Any("error", err))
		return
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to save session", slog.Any("error", err))
	}
}
