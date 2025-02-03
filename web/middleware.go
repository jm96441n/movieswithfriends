package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	fullNameContextKey        = contextKey("fullName")
	currentPartyIDContextKey  = contextKey("currentPartyID")
	emailContextKey           = contextKey("email")
	sessionName               = "moviesWithFriendsCookie"
)

func (a *Application) authenticateMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			logger := a.Logger

			id, err := a.getAccountIDFromSession(req)
			if err != nil {
				if errors.Is(err, ErrFailedToGetAccountIDFromSession) {
					next.ServeHTTP(w, req)
					return
				}
				a.logout(w, req)
				a.setErrorFlashMessage(w, req, "Please log in first.")
				http.Redirect(w, req, "/login", http.StatusInternalServerError)
				return
			}

			exists, err := a.Auth.AccountExists(req.Context(), id)
			if err != nil {
				logger.ErrorContext(ctx, "error fetching id", slog.Any("error", err))
				a.serverError(w, req, err)
				return
			}

			if exists {
				profile, err := a.getProfileFromSession(req)
				if err != nil {
					logger.ErrorContext(ctx, "error fetching profile", slog.Any("error", err))
					a.serverError(w, req, err)
					return
				}

				currentPartyID, err := a.getCurrentPartyIDFromSession(req)
				if err != nil {
					logger.ErrorContext(ctx, "error getting party ID", slog.Any("error", err))
				}

				ctx := context.WithValue(req.Context(), isAuthenticatedContextKey, true)
				ctx = context.WithValue(ctx, emailContextKey, profile.Account.Email)
				ctx = context.WithValue(ctx, fullNameContextKey, profile.FirstName+" "+profile.LastName)

				if currentPartyID > 0 {
					ctx = context.WithValue(ctx, currentPartyIDContextKey, currentPartyID)
				}

				req = req.WithContext(ctx)
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (a *Application) authenticatedMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !isAuthenticated(req.Context()) {
				session, err := a.SessionStore.Get(req, sessionName)
				if err != nil {
					a.serverError(w, req, err)
				}
				session.AddFlash("Please log in first.")
				session.Save(req, w)
				http.Redirect(w, req, "/login", http.StatusSeeOther)
			}

			next.ServeHTTP(w, req)
		})
	}
}

func isAuthenticated(ctx context.Context) bool {
	isAuthenticated := ctx.Value(isAuthenticatedContextKey)
	if isAuthenticated == nil {
		return false
	}
	return isAuthenticated.(bool)
}
