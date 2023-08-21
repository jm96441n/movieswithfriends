package web

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

func loggingMiddlewareBuilder(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cur := time.Now()
			logger.Info(fmt.Sprintf("Starting %s request for %s", req.Method, req.URL.Path))
			next.ServeHTTP(w, req)
			diff := time.Since(cur)
			logger.Info(fmt.Sprintf("Completed %s request for %s in %d milliseconds", req.Method, req.URL.Path, diff.Milliseconds()))
		})
	}
}

func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

			if req.Method != "OPTIONS" {
				next.ServeHTTP(w, req)
			}
		})
	}
}

func SetupWebServer(logger *slog.Logger, router *mux.Router, db *store.PGStore, sessionStore sessions.Store) {
	gob.Register(User{})
	router.HandleFunc("/profile", ProfileShowHandler(logger, db, sessionStore)).Methods("GET", "OPTIONS")
	router.HandleFunc("/signup", SignUpHandler(logger, db)).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", LoginHandler(logger, db, sessionStore)).Methods("POST", "OPTIONS")
	router.Use(loggingMiddlewareBuilder(logger), corsMiddleware())
	// router.HandleFunc("/profile/{id}", ProfileUpdateHandler(logger)).Methods("PUT", "PATCH")
}
