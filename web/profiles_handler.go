package web

import (
	"errors"
	"net/http"

	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find profile in db", "profileID", profileID)
			a.clientError(w, http.StatusNotFound)
			return
		}

		a.Logger.Error("failed to retrieve profile from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewProfilesTemplateData(r, "/profile")
	templateData.Profile = profile
	a.render(w, r, http.StatusOK, "profiles/show.gohtml", templateData)
}
