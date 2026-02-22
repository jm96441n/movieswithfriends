package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/metrics"
)

// LoginShowHandler displays the login form.
func (h *Handlers) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, http.StatusOK, "login/show.gohtml", h.newBaseTemplateData(r, w, "/login"))
}

// LoginHandler processes login form submission and authenticates the user.
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "LoginHandler")
	err := r.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "error parsing form", slog.Any("error", err))
		h.renderer.ServerError(w, r, err)
		return
	}

	profile, err := h.authenticator.Authenticate(r.Context(), logger, r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		if errors.Is(err, identityaccess.ErrInvalidCredentials) {
			h.flashMessager.SetErrorFlashMessage(w, r, "Email/Password combination is incorrect")
			logger.ErrorContext(ctx, "Email/Password combo wrong")

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		logger.ErrorContext(ctx, "error authenticating", slog.Any("error", err))
		h.renderer.ServerError(w, r, err)
		return
	}

	err = h.sessionStore.SetProfileSessionValues(ctx, w, r, profile.Account.ID, profile.ID, profile.FirstName+" "+profile.LastName, profile.Account.Email)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// LogoutHandler logs out the current user.
func (h *Handlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	err := h.Logout(w, r)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}
	h.flashMessager.SetInfoFlashMessage(w, r, "Successfully logged out.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Logout clears the user's session.
// This is exported so that other handlers can log out the user programmatically.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) error {
	_, span, labeler := metrics.SpanFromContext(r.Context(), "identityaccess.http.Handlers.Logout")
	defer span.End()

	err := h.sessionStore.Logout(w, r)
	if err != nil {
		labeler.Add(metrics.ErrorOccurredAttribute())
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
}
