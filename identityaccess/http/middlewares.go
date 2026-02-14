package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/web"
)

type AuthMiddlewares struct {
	Logger          *slog.Logger
	SessionStore    sessionStore
	Auth            *identityaccess.Authenticator
	Renderer        renderer
	ProfilesService *identityaccess.ProfileService
	flashMessager   flashMessager
}

func (a *AuthMiddlewares) AuthenticateMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span, _ := metrics.SpanFromContext(req.Context(), "authenticateMiddleware")
			defer span.End()
			logger := a.Logger

			id, err := a.SessionStore.GetAccountIDFromSession(ctx, req)
			if err != nil {
				if errors.Is(err, web.ErrFailedToGetAccountIDFromSession) {
					next.ServeHTTP(w, req)
					return
				}
				logger.ErrorContext(ctx, "error getting account id from session, logging user out", slog.Any("error", err))
				a.SessionStore.Logout(w, req)
				a.flashMessager.SetErrorFlashMessage(w, req, "Please log in first.")
				http.Redirect(w, req, "/login", http.StatusInternalServerError)
				return
			}

			exists, err := a.Auth.AccountExists(req.Context(), id)
			if err != nil {
				logger.ErrorContext(ctx, "error fetching id", slog.Any("error", err))
				a.Renderer.ServerError(w, req, err)
				return
			}

			if exists {
				profile, err := a.getProfileFromSession(req)
				if err != nil {
					logger.ErrorContext(ctx, "error fetching profile", slog.Any("error", err))
					a.Renderer.ServerError(w, req, err)
					return
				}

				ctx := context.WithValue(req.Context(), web.IsAuthenticatedContextKey, true)
				ctx = context.WithValue(ctx, web.EmailContextKey, profile.Account.Email)
				ctx = context.WithValue(ctx, web.FullNameContextKey, profile.FirstName+" "+profile.LastName)

				req = req.WithContext(ctx)
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (a *AuthMiddlewares) getProfileFromSession(r *http.Request) (*identityaccess.Profile, error) {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.getProfileFromSession")
	defer span.End()
	profileID, err := a.SessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		return nil, err
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("coudl not get profile with id: %d: %w", profileID, err)
	}

	return profile, nil
}

func (a *AuthMiddlewares) AuthenticatedMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span, _ := metrics.SpanFromContext(req.Context(), "authenticatedMiddleware")
			defer span.End()

			if !web.IsAuthenticated(ctx) {
				a.Logger.DebugContext(ctx, "user is not authenticated")
				a.flashMessager.SetErrorFlashMessage(w, req, "Please log in first.")
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
