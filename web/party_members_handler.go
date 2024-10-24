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

	friendID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.Logger.Error("failed to get profile id from session", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	// Add the friend to the party
	err = a.PartyService.AddFriendToParty(ctx, friendID, partyShortID)
	if err != nil {
		a.Logger.Error("failed to add friend to party", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	if r.Header.Get("HX-Request") != "" {
		parties, err := a.PartiesRepository.GetPartiesForMember(ctx, friendID)
		if err != nil {
			a.Logger.Error("failed to get parties for profile", slog.Any("error", err))
			a.serverError(w, r, err)
		}
		templateData := a.NewProfilesTemplateData(r, "/profile")
		templateData.Parties = parties

		a.renderPartial(w, r, http.StatusOK, "profiles/partials/party_list.gohtml", templateData)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
