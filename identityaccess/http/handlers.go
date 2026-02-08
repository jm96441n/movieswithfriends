package http

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
)

const (
	sessionName = "moviesWithFriendsCookie"
)

// Renderer provides template rendering capabilities for HTTP handlers.
// This interface allows domain handlers to render templates without
// depending on the full Application struct.
type Renderer interface {
	Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{})
	SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
	SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string)
	ServerError(w http.ResponseWriter, r *http.Request, err error)
	ClientError(w http.ResponseWriter, r *http.Request, status int, message string)
}

// TemplateDataProvider defines a function that creates template data for rendering.
// This allows the handlers to get template data without depending on the web package.
type TemplateDataProvider func(r *http.Request, w http.ResponseWriter, path string) interface{}

// Handlers holds all HTTP handlers for the identityaccess domain.
// It holds references to domain-specific services and dependencies.
type Handlers struct {
	renderer                 Renderer
	sessionStore             *sessions.CookieStore
	templateDataProvider     TemplateDataProvider
	profileService           *identityaccess.ProfileService
	profileAggregatorService *services.ProfileAggregatorService
	authenticator            *identityaccess.Authenticator
	logger                   *slog.Logger
}

// NewHandlers creates a new Handlers instance with the required dependencies.
func NewHandlers(
	renderer Renderer,
	sessionStore *sessions.CookieStore,
	templateDataProvider TemplateDataProvider,
	profileService *identityaccess.ProfileService,
	profileAggregatorService *services.ProfileAggregatorService,
	authenticator *identityaccess.Authenticator,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		renderer:                 renderer,
		sessionStore:             sessionStore,
		templateDataProvider:     templateDataProvider,
		profileService:           profileService,
		profileAggregatorService: profileAggregatorService,
		authenticator:            authenticator,
		logger:                   logger,
	}
}
