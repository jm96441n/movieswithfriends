package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/web"
)

var ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")

// Handlers holds all HTTP handlers for the partymgmt domain.
// It embeds web.Renderer for template rendering and flash messages,
// and contains domain-specific services for business logic.
type Handlers struct {
	renderer                 renderer
	logger                   *slog.Logger
	flashMessager            flashMessager
	sessionStore             SessionStore
	newBaseTemplateData      func(*http.Request, http.ResponseWriter, string) web.BaseTemplateData
	moviesService            *partymgmt.MovieService
	partyService             partymgmt.PartyService
	watcherService           partymgmt.WatcherService
	invitationsService       partymgmt.InvitationsService
	partiesRepository        store.PartyRepository
	profileAggregatorService *services.ProfileAggregatorService
	telemetry                metrics.TelemetryProvider
}

type renderer interface {
	Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	ServerError(w http.ResponseWriter, r *http.Request, err error)
	ClientError(w http.ResponseWriter, r *http.Request, status int, message string)
}

type SessionStore interface {
	GetProfileIDFromSession(ctx context.Context, r *http.Request) (int, error)
	Logout(w http.ResponseWriter, r *http.Request) error
}

type flashMessager interface {
	SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
	SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
}

// Config contains the dependencies required to create Handlers.
type Config struct {
	Renderer                 renderer
	Logger                   *slog.Logger
	FlashMessager            flashMessager
	SessionStore             SessionStore
	NewBaseTemplateData      func(*http.Request, http.ResponseWriter, string) web.BaseTemplateData
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
		renderer:                 cfg.Renderer,
		logger:                   cfg.Logger,
		flashMessager:            cfg.FlashMessager,
		sessionStore:             cfg.SessionStore,
		newBaseTemplateData:      cfg.NewBaseTemplateData,
		moviesService:            cfg.MoviesService,
		partyService:             cfg.PartyService,
		watcherService:           cfg.WatcherService,
		invitationsService:       cfg.InvitationsService,
		partiesRepository:        cfg.PartiesRepository,
		profileAggregatorService: cfg.ProfileAggregatorService,
		telemetry:                cfg.Telemetry,
	}
}

// getWatcherFromSession retrieves the watcher from the session.
func (h *Handlers) getWatcherFromSession(ctx context.Context, r *http.Request) (partymgmt.Watcher, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Handlers.getWatcherFromSession")
	defer span.End()

	watcherID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
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
	h.flashMessager.SetErrorFlashMessage(w, r, "There was an issue authenticating you, please log in again.")
	err = h.sessionStore.Logout(w, r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to logout user after failed session retrieval", slog.Any("error", err))
	}
	http.Redirect(w, r, "/login", http.StatusInternalServerError)
}
