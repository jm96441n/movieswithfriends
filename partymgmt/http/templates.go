package http

import (
	"net/http"

	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/web"
)

type AddMovieToPartiesModalTemplateData struct {
	MovieID         int
	TMDBID          int
	AddedParties    []partymgmt.Party
	NotAddedParties []partymgmt.Party
}

type MoviesTemplateData struct {
	Movies                   []partymgmt.TMDBMovie
	Movie                    partymgmt.Movie
	MovieAddedToCurrentParty bool
	SearchValue              string
	web.BaseTemplateData
}

// InviteModalTemplateData holds the data for rendering the invite modal.
type InviteModalTemplateData struct {
	CreateErrorMsg string
	FetchErrorMsg  string
	PendingInvites []partymgmt.Invite
	PartyID        int
	ShowModal      bool
}

type PartiesTemplateData struct {
	Party                 partymgmt.Party
	CurrentWatcherIsOwner bool
	Members               []partymgmt.PartyMember
	ModalData             InviteModalTemplateData
	web.BaseTemplateData
}

type PartiesIndexTemplateData struct {
	Parties        []partymgmt.Party
	InvitedParties []partymgmt.Party
	CurrentUserID  int
	web.BaseTemplateData
}

type WatchListTemplateData struct {
	WatchedMovies []partymgmt.PartyMovie
	NumPages      int
	CurPage       int
	web.BaseTemplateData
}

func (h *Handlers) NewMoviesTemplateData(r *http.Request, w http.ResponseWriter, path string) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}

func (h *Handlers) NewPartiesIndexTemplateData(r *http.Request, w http.ResponseWriter, path string, parties, invitedParties []partymgmt.Party, currentUserID int) PartiesIndexTemplateData {
	return PartiesIndexTemplateData{
		Parties:          parties,
		InvitedParties:   invitedParties,
		CurrentUserID:    currentUserID,
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}

func (h *Handlers) NewPartiesTemplateData(r *http.Request, w http.ResponseWriter, path string) PartiesTemplateData {
	return PartiesTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}

func (h *Handlers) NewWatchListTemplateData(r *http.Request, w http.ResponseWriter, path string) WatchListTemplateData {
	return WatchListTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}
