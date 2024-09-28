package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParam := r.PathValue("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		a.clientError(w, http.StatusBadRequest)
		return
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find profoile in db", "id", id)
			a.clientError(w, http.StatusNotFound)
			return
		}

		a.Logger.Error("failed to retrieve profile from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewProfilesTemplateData(r, "/profiles/1")
	templateData.Profile = profile
	a.render(w, r, http.StatusOK, "profiles/show.gohtml", templateData)
}
