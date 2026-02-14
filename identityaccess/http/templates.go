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

func NewProfilesTemplateData(r *http.Request, w http.ResponseWriter, currentPath string, renderer renderer) *ProfilesTemplateData {
	return &ProfilesTemplateData{
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, currentPath),
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

func NewSignupTemplateData(r *http.Request, w http.ResponseWriter, path string, renderer renderer) *SignupTemplateData {
	return &SignupTemplateData{
		BaseTemplateData: renderer.NewBaseTemplateData(r, w, path),
	}
}
