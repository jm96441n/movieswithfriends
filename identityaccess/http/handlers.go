package http

import (
	"log/slog"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/web"
)

// Handlers holds all HTTP handlers for the identityaccess domain.
// It embeds the web.Renderer interface to provide template rendering capabilities
// and holds references to domain-specific services and dependencies.
type Handlers struct {
	renderer                 web.Renderer
	profileService           *identityaccess.ProfileService
	profileAggregatorService *services.ProfileAggregatorService
	authenticator            *identityaccess.Authenticator
	logger                   *slog.Logger
}

// NewHandlers creates a new Handlers instance with the required dependencies.
func NewHandlers(
	renderer web.Renderer,
	profileService *identityaccess.ProfileService,
	profileAggregatorService *services.ProfileAggregatorService,
	authenticator *identityaccess.Authenticator,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		renderer:                 renderer,
		profileService:           profileService,
		profileAggregatorService: profileAggregatorService,
		authenticator:            authenticator,
		logger:                   logger,
	}
}
