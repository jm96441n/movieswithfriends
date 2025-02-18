package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/metrics"
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
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "CreateInviteHandler")
	defer span.End()
	logger := a.Logger.With("handler", "InvitationsHandler")

	partyID, email, err := parseInviteForm(r)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to parse form", slog.Any("error", err))
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
		logger.ErrorContext(ctx, "Failed to create invite", slog.Any("error", err))
		templateData.CreateErrorMsg = "There was an error inviting this member, try again."
	}

	invited, err := a.InvitationsService.GetInvitationsForParty(ctx, partyID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get invitations", slog.Any("error", err))
		templateData.FetchErrorMsg = "There was an error loading pending invites."
	} else {
		templateData.PendingInvites = invited
	}

	logger.InfoContext(ctx, "successfully invited user")

	a.renderPartial(w, r, http.StatusOK, "parties/partials/invite_modal.gohtml", templateData)
}

func (a *Application) AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	// ctx, span, _ := metrics.SpanFromContext(r.Context(), "AcceptInviteHandler")
	// defer span.End()
	// logger := a.Logger.With("handler", "InvitationsHandler")

	// this should remove the invitation, create the party_member record
	// and then cause a re-render of the parties listing
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
