package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/metrics"
)

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	fullNameContextKey        = contextKey("fullName")
	currentPartyIDContextKey  = contextKey("currentPartyID")
	emailContextKey           = contextKey("email")
	sessionName               = "moviesWithFriendsCookie"
)

func (a *Application) GetLogger(ctx context.Context) *slog.Logger {
	// span := trace.SpanFromContext(ctx)
	return a.Logger //.With("trace.trace_id", span.SpanContext().TraceID().String(), "trace.span_id", span.SpanContext().SpanID().String())
}

func (a *Application) authenticateMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span, _ := metrics.SpanFromContext(req.Context(), "authenticateMiddleware")
			defer span.End()
			logger := a.GetLogger(ctx)

			logger.InfoContext(ctx, "checking if user is authenticated")
			id, err := a.getAccountIDFromSession(ctx, req)
			if err != nil {
				if errors.Is(err, ErrFailedToGetAccountIDFromSession) {
					next.ServeHTTP(w, req)
					return
				}
				logger.ErrorContext(ctx, "error getting account id from session, logging user out", slog.Any("error", err))
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

				ctx := context.WithValue(req.Context(), isAuthenticatedContextKey, true)
				ctx = context.WithValue(ctx, emailContextKey, profile.Account.Email)
				ctx = context.WithValue(ctx, fullNameContextKey, profile.FirstName+" "+profile.LastName)

				req = req.WithContext(ctx)
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (a *Application) authenticatedMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, span, _ := metrics.SpanFromContext(req.Context(), "authenticatedMiddleware")
			defer span.End()

			if !isAuthenticated(ctx) {
				a.GetLogger(ctx).ErrorContext(ctx, "user is not authenticated")
				a.setErrorFlashMessage(w, req, "Please log in first.")
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func isAuthenticated(ctx context.Context) bool {
	ctx, span, _ := metrics.SpanFromContext(ctx, "isAuthenticated")
	defer span.End()
	isAuthenticated := ctx.Value(isAuthenticatedContextKey)
	if isAuthenticated == nil {
		return false
	}
	return isAuthenticated.(bool)
}
