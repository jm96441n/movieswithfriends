package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/partymgmt"
)

type InviteModalTemplateData struct {
	CreateErrorMsg string
	FetchErrorMsg  string
	PendingInvites []partymgmt.Invite
	PartyID        int
	ShowModal      bool
}

func (a *Application) CreateInviteHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "InvitationsHandler")
	ctx := r.Context()

	partyID, email, err := parseInviteForm(r)
	if err != nil {
		logger.Error("Failed to parse form", slog.Any("error", err))
		a.setErrorFlashMessage(w, r, "There was an error inviting this user, try again.")
		w.Header().Set("HX-Redirect", "/parties")
		return
	}

	templateData := InviteModalTemplateData{
		PartyID:   partyID,
		ShowModal: true,
	}
	err = a.InvitationsService.CreateInvite(ctx, a.WatcherService, partyID, email)
	if err != nil {
		logger.Error("Failed to create invite", slog.Any("error", err))
		templateData.CreateErrorMsg = "There was an error inviting this member, try again."
	}

	invited, err := a.InvitationsService.GetInvitationsForParty(ctx, partyID)
	if err != nil {
		logger.Error("Failed to get invitations", slog.Any("error", err))
		templateData.FetchErrorMsg = "There was an error loading pending invites."
	} else {
		templateData.PendingInvites = invited
	}

	logger.Info("successfully invited user")

	a.renderPartial(w, r, http.StatusOK, "parties/partials/invite_modal.gohtml", templateData)
}

func parseInviteForm(r *http.Request) (int, string, error) {
	err := r.ParseForm()
	if err != nil {
		return 0, "", err
	}

	partyID, err := strconv.Atoi(r.FormValue("partyID"))
	if err != nil {
		return 0, "", err
	}

	email := r.FormValue("email")

	return partyID, email, nil
}
