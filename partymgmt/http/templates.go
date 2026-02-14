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

func NewMoviesTemplateData(r *http.Request, w http.ResponseWriter, path string, renderer renderer) MoviesTemplateData {
	return MoviesTemplateData{
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, path),
	}
}

func NewPartiesIndexTemplateData(r *http.Request, w http.ResponseWriter, path string, renderer renderer, parties, invitedParties []partymgmt.Party, currentUserID int) PartiesIndexTemplateData {
	return PartiesIndexTemplateData{
		Parties:          parties,
		InvitedParties:   invitedParties,
		CurrentUserID:    currentUserID,
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, path),
	}
}

func NewPartiesTemplateData(r *http.Request, w http.ResponseWriter, path string, renderer renderer) PartiesTemplateData {
	return PartiesTemplateData{
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, path),
	}
}

func NewWatchListTemplateData(r *http.Request, w http.ResponseWriter, path string, renderer renderer) WatchListTemplateData {
	return WatchListTemplateData{
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, path),
	}
}
