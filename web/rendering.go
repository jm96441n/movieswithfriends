package web

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/jm96441n/movieswithfriends/metrics"
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

// Render renders a full page template with the base layout.
func (a *Application) Render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.templateCache[page]

	if !ok {
		a.ServerError(w, r, fmt.Errorf("template does not exist for page %q", page))
		return
	}

	buf := bytes.NewBuffer([]byte{})

	// write to the buffer to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// RenderPartial renders a partial template without the base layout.
// This is typically used for HTMX responses.
func (a *Application) RenderPartial(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	ts, ok := a.templateCache[page]

	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.RenderPartial")
	defer span.End()

	if !ok {
		a.Logger.ErrorContext(ctx, "template does not exist for page", slog.Any("page", page))
		a.ServerError(w, r, fmt.Errorf("template does not exist for page %q", page))
		return
	}

	tmplName := filepath.Base(page)

	buf := bytes.NewBuffer([]byte{})
	// write to the buffer to check if we have any template errors before rendering the response
	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		a.ServerError(w, r, err)
		return
	}

	a.Logger.InfoContext(ctx, "rendering template", slog.Any("template", page))

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// SetInfoFlashMessage sets an info flash message in the session.
func (a *Application) SetInfoFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	_, span, _ := metrics.SpanFromContext(r.Context(), "Application.SetInfoFlashMessage")
	defer span.End()

	a.SetFlashMessage(w, r, FlashInfoKey, msg)
}

// SetErrorFlashMessage sets an error flash message in the session.
func (a *Application) SetErrorFlashMessage(w http.ResponseWriter, r *http.Request, msg string) {
	_, span, _ := metrics.SpanFromContext(r.Context(), "Application.SetErrorFlashMessage")
	defer span.End()

	a.SetFlashMessage(w, r, FlashErrorKey, msg)
}

// SetFlashMessage sets a flash message in the session with the specified key.
func (a *Application) SetFlashMessage(w http.ResponseWriter, r *http.Request, key, msg string) {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Application.SetFlashMessage")
	defer span.End()

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.Logger.ErrorContext(ctx, "failed to get session", slog.Any("error", err))
		return
	}

	a.Logger.DebugContext(ctx, "adding flash message")
	session.AddFlash(msg, key)
	session.Save(r, w)
}

// ServerError logs the error and returns a 500 Internal Server Error response.
func (a *Application) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	ctx, span, _ := metrics.SpanFromContext(r.Context(), "serverError")
	defer span.End()

	a.Logger.ErrorContext(ctx, err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ClientError sets an error flash message and returns the specified HTTP status code.
func (a *Application) ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.ServerError(w, r, err)
		return
	}
	session.AddFlash(message)
	session.Save(r, w)

	http.Error(w, http.StatusText(status), status)
}
