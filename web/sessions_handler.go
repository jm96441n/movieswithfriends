package web

import (
	"errors"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/login")
	a.render(w, r, http.StatusOK, "login/show.gohtml", data)
}

func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.Logger.Error("error parsing form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	account, err := a.AccountService.FindAccountByEmail(r.Context(), r.FormValue("email"))
	if err != nil {
		a.Logger.Error("error finding account by email", slog.Any("error", err), slog.String("email", r.FormValue("email")))
		data := a.NewTemplateData(r, "/signup")
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
		return
	}

	err = bcrypt.CompareHashAndPassword(account.Password, []byte(r.PostForm.Get("password")))
	if err != nil {
		data := a.NewTemplateData(r, "/signup")
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			a.Logger.Error("error comparing password", slog.Any("error", err))
			a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
			return
		}
		a.Logger.Error("error comparing password", slog.Any("error", err))
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
		return
	}
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	session.Values["accountID"] = account.ID
	session.Values["profileID"] = account.Profile.ID

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
