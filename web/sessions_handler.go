package web

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

	profile, err := a.Auth.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		if errors.Is(err, identityaccess.ErrInvalidCredentials) {
			a.setErrorFlashMessage(w, r, "Email/Password combination is incorrect")
			a.Logger.Error("Email/Password combo wrong")

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		a.Logger.Error("error authenticating", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.Logger.Error("error getting session from store success path", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	// TODO: use ACL to translate betwwen profile and watcher
	watcher, _ := a.WatcherService.GetWatcher(r.Context(), profile.ID)
	currentPartyID, err := watcher.GetCurrentPartyID(r.Context())
	if err != nil {
		a.Logger.Error("error getting session from store success path", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	session.Values["accountID"] = profile.Account.ID
	session.Values["profileID"] = profile.ID
	session.Values["fullName"] = profile.FirstName + " " + profile.LastName
	session.Values["email"] = profile.Account.Email
	a.setCurrentPartyInSession(r, w, currentPartyID)

	err = session.Save(r, w)
	if err != nil {
		a.Logger.Error("error saving session", slog.Any("error", err))
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
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *Application) logout(w http.ResponseWriter, r *http.Request) error {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}

func (a *Application) SetCurrentPartyHandler(w http.ResponseWriter, r *http.Request) {
	idPartyParam := r.PathValue("party_id")
	// TODO: make sure party is a valid party id for the given user
	idParty, err := strconv.Atoi(idPartyParam)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	currentPartyID, err := a.getCurrentPartyIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	err = a.setCurrentPartyInSession(r, w, idParty)
	if err != nil {
		a.clientError(w, r, http.StatusBadRequest, "uh oh")
		return
	}

	if idParty != currentPartyID {
		w.Header().Set("HX-Trigger", "changeCurrentParty")
	}

	data := a.NewSidebarTemplateData(r, w, idParty)
	a.renderPartial(w, r, http.StatusOK, "partials/sidebar_parties.gohtml", data)
}

func (a *Application) GetSidebarParties(w http.ResponseWriter, r *http.Request) {
	currentPartyID, err := a.getCurrentPartyIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	data := a.NewSidebarTemplateData(r, w, currentPartyID)
	a.renderPartial(w, r, http.StatusOK, "partials/sidebar_parties.gohtml", data)
}

func (a *Application) setCurrentPartyInSession(r *http.Request, w http.ResponseWriter, idParty int) error {
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	session.Values["currentPartyID"] = idParty

	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}
