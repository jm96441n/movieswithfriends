package web

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/metrics"
)

func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, w, "/login")
	a.render(w, r, http.StatusOK, "login/show.gohtml", data)
}

func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.GetLogger(ctx).With("handler", "LoginHandler")
	err := r.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "error parsing form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	profile, err := a.Auth.Authenticate(r.Context(), logger, r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		if errors.Is(err, identityaccess.ErrInvalidCredentials) {
			a.setErrorFlashMessage(w, r, "Email/Password combination is incorrect")
			logger.ErrorContext(ctx, "Email/Password combo wrong")

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		logger.ErrorContext(ctx, "error authenticating", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		logger.ErrorContext(ctx, "error getting session from store success path", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	session.Values["accountID"] = profile.Account.ID
	session.Values["profileID"] = profile.ID
	session.Values["fullName"] = profile.FirstName + " " + profile.LastName
	session.Values["email"] = profile.Account.Email

	err = session.Save(r, w)
	if err != nil {
		logger.ErrorContext(ctx, "error saving session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (a *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	err := a.logout(w, r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.setInfoFlashMessage(w, r, "Successfully logged out.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *Application) logout(w http.ResponseWriter, r *http.Request) error {
	_, span, labeler := metrics.SpanFromContext(r.Context(), "web.Application.logout")
	defer span.End()

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		labeler.Add(metrics.ErrorOccurredAttribute())
		return err
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		labeler.Add(metrics.ErrorOccurredAttribute())
		return err
	}
	return nil
}
