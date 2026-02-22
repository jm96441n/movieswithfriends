package http

import (
	"net/http"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/web"
)

type ProfilesTemplateData struct {
	Profile           *identityaccess.Profile
	WatchedMovies     []partymgmt.PartyMovie
	Parties           []partymgmt.Party
	InvitedParties    []partymgmt.Party
	CurrentUserID     int
	NumPages          int
	CurPage           int
	HasEmailError     *bool
	HasPasswordError  *bool
	HasFirstNameError *bool
	HasLastNameError  *bool
	web.BaseTemplateData
}

func (h *Handlers) NewProfilesTemplateData(r *http.Request, w http.ResponseWriter, currentPath string) *ProfilesTemplateData {
	return &ProfilesTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, currentPath),
	}
}

func (s *ProfilesTemplateData) InitHasErrorFields() {
	s.HasEmailError = new(bool)
	s.HasPasswordError = new(bool)
	s.HasFirstNameError = new(bool)
	s.HasLastNameError = new(bool)
}

func (s *ProfilesTemplateData) SetProfile(profile *identityaccess.Profile) {
	s.Profile = profile
}

func (s *ProfilesTemplateData) SetParties(parties, invitedParties interface{}) {
	if p, ok := parties.([]partymgmt.Party); ok {
		s.Parties = p
	}
	if ip, ok := invitedParties.([]partymgmt.Party); ok {
		s.InvitedParties = ip
	}
}

func (s *ProfilesTemplateData) SetWatchedMovies(movies interface{}) {
	if m, ok := movies.([]partymgmt.PartyMovie); ok {
		s.WatchedMovies = m
	}
}

func (s *ProfilesTemplateData) SetPagination(curPage, numPages int) {
	s.CurPage = curPage
	s.NumPages = numPages
}

func (s *ProfilesTemplateData) GetErrorFields() (email, password, firstName, lastName **bool) {
	return &s.HasEmailError, &s.HasPasswordError, &s.HasFirstNameError, &s.HasLastNameError
}

type SignupTemplateData struct {
	HasEmailError     *bool
	HasPasswordError  *bool
	HasFirstNameError *bool
	HasLastNameError  *bool
	web.BaseTemplateData
}

func (s *SignupTemplateData) InitHasErrorFields() {
	s.HasEmailError = new(bool)
	s.HasPasswordError = new(bool)
	s.HasFirstNameError = new(bool)
	s.HasLastNameError = new(bool)
}

func (s *SignupTemplateData) GetErrorFields() (email, password, firstName, lastName **bool) {
	return &s.HasEmailError, &s.HasPasswordError, &s.HasFirstNameError, &s.HasLastNameError
}

func (h *Handlers) NewSignupTemplateData(r *http.Request, w http.ResponseWriter, path string) *SignupTemplateData {
	return &SignupTemplateData{
		BaseTemplateData: h.newBaseTemplateData(r, w, path),
	}
}
