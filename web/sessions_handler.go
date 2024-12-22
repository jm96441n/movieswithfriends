package web

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
)

func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, w, "/login")
	a.render(w, r, http.StatusOK, "login/show.gohtml", data)
}

func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.Logger.Error("error parsing form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	account, err := a.Auth.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		if errors.Is(err, identityaccess.ErrInvalidCredentials) {
			a.clientError(w, http.StatusUnauthorized)
			return
		}
		a.serverError(w, r, err)
		return
	}

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	session.Values["accountID"] = account.ID
	session.Values["profileID"] = account.Profile.ID
	session.Values["fullName"] = account.Profile.FirstName + " " + account.Profile.LastName
	session.Values["email"] = account.Email

	err = session.Save(r, w)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (a *Application) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
