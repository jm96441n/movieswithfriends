package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/web"
)

type flashMessager interface {
	SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
	SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
}

type renderer interface {
	NewBaseTemplateData(r *http.Request, w http.ResponseWriter, path string) web.BaseTemplateData
	Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	ServerError(w http.ResponseWriter, r *http.Request, err error)
	ClientError(w http.ResponseWriter, r *http.Request, status int, message string)
}

type sessionStore interface {
	GetProfileIDFromSession(ctx context.Context, r *http.Request) (int, error)
	GetAccountIDFromSession(ctx context.Context, r *http.Request) (int, error)
	SetProfileSessionValues(ctx context.Context, w http.ResponseWriter, r *http.Request, accountID, profileID int, fullName, email string) error
	Logout(w http.ResponseWriter, r *http.Request) error
}

// Handlers holds all HTTP handlers for the identityaccess domain.
// It holds references to domain-specific services and dependencies.
type Handlers struct {
	renderer                 renderer
	flashMessager            flashMessager
	sessionStore             sessionStore
	profileService           *identityaccess.ProfileService
	profileAggregatorService *services.ProfileAggregatorService
	authenticator            *identityaccess.Authenticator
	telemetry                metrics.TelemetryProvider
	logger                   *slog.Logger
}

// NewHandlers creates a new Handlers instance with the required dependencies.
func NewHandlers(
	renderer renderer,
	flashMessager flashMessager,
	sessionStore sessionStore,
	profileService *identityaccess.ProfileService,
	profileAggregatorService *services.ProfileAggregatorService,
	authenticator *identityaccess.Authenticator,
	telemetry metrics.TelemetryProvider,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		renderer:                 renderer,
		flashMessager:            flashMessager,
		sessionStore:             sessionStore,
		profileService:           profileService,
		profileAggregatorService: profileAggregatorService,
		authenticator:            authenticator,
		telemetry:                telemetry,
		logger:                   logger,
	}
}
