package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, profileID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find profile in db", "profileID", profileID)
			status = http.StatusNotFound
		}

		a.setErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		a.logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	parties, err := a.PartiesRepository.GetPartiesForMember(ctx, profileID)
	if err != nil {
		a.Logger.Error("failed to retrieve parties from db", "error", err)
		a.serverError(w, r, err)
	}

	watchedMovies, numMovies, err := a.MemberService.GetWatchHistory(ctx, profileID, 0)
	if err != nil {
		a.Logger.Error("failed to retrieve watched movies from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	numPages := numMovies / 5
	if numMovies > numPages*5 {
		numPages++
	}

	templateData := a.NewProfilesTemplateData(r, w, "/profile")
	templateData.Profile = profile
	templateData.Parties = parties
	templateData.WatchedMovies = watchedMovies
	templateData.CurPage = 1
	templateData.NumPages = numPages
	a.render(w, r, http.StatusOK, "profiles/show.gohtml", templateData)
}

func (a *Application) ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	profile, err := a.ProfilesService.GetProfileByID(ctx, profileID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find profile in db", "profileID", profileID)
			status = http.StatusNotFound
		}

		a.setErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		a.logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	templateData := a.NewProfilesTemplateData(r, w, "/profile")
	templateData.Profile = profile
	a.render(w, r, http.StatusOK, "profiles/edit.gohtml", templateData)
}

func (a *Application) GetPaginatedWatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "GetPaginatedWatchHistoryHandler")
	logger.Debug("getting paginated movies list")

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	page := r.URL.Query().Get("page")

	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	offset := 5 * (pageNum - 1)

	watchedMovies, numMovies, err := a.MemberService.GetWatchHistory(ctx, profileID, offset)
	if err != nil {
		a.Logger.Error("failed to retrieve watched movies from db", "error", err)
		a.serverError(w, r, err)
		return
	}

	numPages := numMovies / 5
	if numMovies > numPages*5 {
		numPages++
	}

	templateData := a.NewProfilesTemplateData(r, w, "/profile")
	templateData.WatchedMovies = watchedMovies
	templateData.CurPage = pageNum
	templateData.NumPages = numPages
	a.renderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
}
