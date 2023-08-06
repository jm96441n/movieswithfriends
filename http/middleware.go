package http

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

func loggingMiddlewareBuilder(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cur := time.Now()
			logger.Info(fmt.Sprintf("Starting request for %s", req.URL.Path))
			next.ServeHTTP(w, req)
			diff := time.Since(cur)
			logger.Info("Completed request for %s in %d milliseconds", req.URL.Path, diff.Milliseconds())
		})
	}
}

func SetupWebServer(logger *slog.Logger, router *mux.Router, templates map[string]*template.Template) {
	router.HandleFunc("/profiles/{id}", ProfileShowHandler(logger, templates[profileShowKey])).Methods("GET")
	router.Use(loggingMiddlewareBuilder(logger))
	// router.HandleFunc("/profile/{id}", ProfileUpdateHandler(logger)).Methods("PUT", "PATCH")
}
