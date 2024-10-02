package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	sessionName               = "moviesWithFriendsCookie"
)

func loggingMiddlewareBuilder(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cur := time.Now()
			logger.Info(fmt.Sprintf("Starting %s request for %s", req.Method, req.URL.Path))
			next.ServeHTTP(w, req)
			diff := time.Since(cur)
			logger.Info(fmt.Sprintf("Completed %s request for %s in %d milliseconds", req.Method, req.URL.Path, diff.Milliseconds()))
		})
	}
}

func (a *Application) authenticateMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			session, err := a.SessionStore.Get(req, sessionName)
			if err != nil {
				a.Logger.Error("failed to get session", slog.Any("error", err))
				a.serverError(w, req, err)
				return
			}

			accountID := session.Values["accountID"]

			if accountID == nil {
				a.Logger.Error("no accountID in session")
				next.ServeHTTP(w, req)
				return
			}

			id := accountID.(int)

			exists, err := a.AccountService.AccountExists(req.Context(), id)
			if err != nil {
				a.Logger.Error("error fetching id", slog.Any("error", err))
				a.serverError(w, req, err)
				return
			}

			if exists {
				ctx := context.WithValue(req.Context(), isAuthenticatedContextKey, true)
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
				a.clientError(w, http.StatusUnauthorized)
				return
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

func corsMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4000")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

			if req.Method != "OPTIONS" {
				next.ServeHTTP(w, req)
			}
		})
	}
}
