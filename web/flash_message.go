package web

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/metrics"
)

type FlashMessager struct {
	logger       *slog.Logger
	sessionStore *sessions.CookieStore
}

func NewFlashMessager(logger *slog.Logger, sessionStore *sessions.CookieStore) *FlashMessager {
	// TODO: refactor to depend on sesssion accessor
	return &FlashMessager{
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (a *FlashMessager) SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	_, span, _ := metrics.SpanFromContext(r.Context(), "Application.setInfoFlashMessage")
	defer span.End()

	a.setFlashMessage(w, r, FlashInfoKey, msg)
}

func (a *FlashMessager) SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	_, span, _ := metrics.SpanFromContext(r.Context(), "Application.setErrorFlashMessage")
	defer span.End()
	a.setFlashMessage(w, r, FlashErrorKey, msg)
}

func (a *FlashMessager) setFlashMessage(w http.ResponseWriter, r *http.Request, key, msg string) {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.setFlashMessage")
	defer span.End()

	session, err := a.sessionStore.Get(r, sessionName)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to get session", slog.Any("error", err))
		return
	}

	a.logger.DebugContext(ctx, "adding flash message")
	session.AddFlash(msg, key)
	session.Save(r, w)
}
