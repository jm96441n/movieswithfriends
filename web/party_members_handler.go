package web

import (
	"log/slog"
	"net/http"
)

func (a *Application) AddMemberToPartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		a.Logger.Error("failed to parse form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	partyShortID := r.FormValue("party_short_id")

	watcher, err := a.getWatcherFromSession(r)
	if err != nil {
		a.Logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	party, err := a.PartyService.GetPartyByShortID(ctx, partyShortID)
	if err != nil {
		a.Logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	// Add the friend to the party
	err = party.AddMember(ctx, watcher.ID)
	if err != nil {
		a.Logger.Error("failed to add friend to party", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	if r.Header.Get("HX-Request") != "" {
		parties, err := watcher.GetParties(ctx)
		if err != nil {
			a.Logger.Error("failed to get parties for profile", slog.Any("error", err))
			a.serverError(w, r, err)
		}
		templateData := a.NewProfilesTemplateData(r, w, "/profile")
		templateData.Parties = parties

		a.renderPartial(w, r, http.StatusOK, "profiles/partials/party_list.gohtml", templateData)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
