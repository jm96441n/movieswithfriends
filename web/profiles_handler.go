package web

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/store"
)

func (a *Application) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := a.Logger.With("handler", "ProfileShowHandler")
	logger.Debug("getting profile info")

	profileID, err := a.getProfileIDFromSession(r)
	if err != nil {
		logger.Error("failed to get profile id from session", "error", err)
		a.setErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		a.logout(w, r)
		http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	pageData, err := a.ProfileAggregatorService.GetProfilePageData(ctx, logger, profileID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			logger.Error("did not find profile in db", "profileID", profileID)
			status = http.StatusNotFound
		} else {
			logger.Error("failed to retrieve profile from db", "error", err)
		}

		a.setErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		a.logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}
	templateData := a.NewProfilesTemplateData(r, w, "/profile")
	templateData.Profile = pageData.Profile
	templateData.Parties = pageData.Parties
	templateData.WatchedMovies = pageData.WatchedMovies
	templateData.CurPage = pageData.CurPage
	templateData.NumPages = pageData.NumPages
	logger.Info("successfully loaded profile info")
	a.render(w, r, http.StatusOK, "profiles/show.gohtml", templateData)
}

func (a *Application) ProfileEditPageHandler(w http.ResponseWriter, r *http.Request) {
	// logger := a.Logger.With("handler", "ProfileEditPageHandler")

	profile, err := a.getProfileFromSession(r)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("did not find profile in db", "profileID", profile.ID)
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

func (a *Application) ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.Logger.With("handler", "ProfileEditHandler")
	ctx := r.Context()

	profile, err := a.getProfileFromSession(r)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			logger.Error("did not find profile in db", slog.Any("error", err))
			status = http.StatusNotFound
		}

		a.setErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		a.logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	req, err := parseEditProfileForm(r)
	if err != nil {
		templateData := a.NewProfilesTemplateData(r, w, "/profile")
		templateData.Profile = profile
		a.setErrorFlashMessage(w, r, "There was an error editing your profile, please try again")

		a.render(w, r, http.StatusBadRequest, "profiles/edit.gohtml", templateData)
		return
	}

	logger.Info("req to update profile", "req", req)

	err = profile.Update(ctx, logger, req)
	if err != nil {
		templateData := a.NewProfilesTemplateData(r, w, "/profile")
		templateData.Profile = profile

		var editErr *identityaccess.ProfileEditValidationError

		a.setErrorFlashMessage(w, r, "There was an error editing your profile, please try again")
		if errors.As(err, &editErr) {
			templateData.InitHasErrorFields()

			if editErr.EmailError != nil {
				*templateData.HasEmailError = true
			}

			if editErr.PasswordError != nil {
				*templateData.HasPasswordError = true
			}

			if editErr.FirstNameError != nil {
				*templateData.HasFirstNameError = true
			}

			if editErr.LastNameError != nil {
				*templateData.HasLastNameError = true
			}
		}

		a.render(w, r, http.StatusBadRequest, "profiles/edit.gohtml", templateData)
		return
	}

	a.setInfoFlashMessage(w, r, "Edited your profile!")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func parseEditProfileForm(r *http.Request) (identityaccess.ProfileUpdateReq, error) {
	r.ParseForm()
	req := identityaccess.ProfileUpdateReq{
		FirstName:               r.FormValue("firstName"),
		LastName:                r.FormValue("lastName"),
		Email:                   r.FormValue("email"),
		CurrentPassword:         r.FormValue("currentPassword"),
		NewPassword:             r.FormValue("newPassword"),
		NewPasswordConfirmation: r.FormValue("confirmPassword"),
	}

	return req, nil
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

	movieData, err := a.ProfileAggregatorService.GetWatchPaginatedHistory(ctx, logger, profileID, pageNum)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	templateData := a.NewProfilesTemplateData(r, w, "/profile")
	templateData.WatchedMovies = movieData.WatchedMovies
	templateData.CurPage = pageNum
	templateData.NumPages = movieData.NumPages
	a.renderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
}
