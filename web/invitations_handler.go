package web

import (
	"net/http"

	"github.com/jm96441n/movieswithfriends/partymgmt"
)

type InviteModalTemplateData struct {
	ErrorMsg       string
	PendingInvites []partymgmt.Invite
}

func (a *Application) CreateInviteHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "InvitationsHandler")
	logger.Info("calling InvitationsHandler")
	templateData := InviteModalTemplateData{}
	a.renderPartial(w, r, http.StatusOK, "parties/partials/invite_modal.gohtml", templateData)
}
