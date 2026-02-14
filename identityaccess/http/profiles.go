package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/metrics"
)

// ProfileShowHandler displays the user's profile page with their parties and watch history.
func (h *Handlers) ProfileShowHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "ProfileShowHandler")
	logger.DebugContext(ctx, "getting profile info")

	profileID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get profile id from session", "error", err)
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		h.Logout(w, r)
		http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	pageData, err := h.profileAggregatorService.GetProfilePageData(ctx, logger, profileID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			logger.ErrorContext(ctx, "did not find profile in db", "profileID", profileID)
			status = http.StatusNotFound
		} else {
			logger.ErrorContext(ctx, "failed to retrieve profile from db", "error", err)
		}

		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		h.Logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	// Get profile-specific template data
	templateData := NewProfilesTemplateData(r, w, "/profile", h.renderer)
	templateData.SetProfile(pageData.Profile)
	templateData.SetParties(pageData.Parties, pageData.InvitedParties)
	templateData.SetWatchedMovies(pageData.WatchedMovies)
	templateData.SetPagination(pageData.CurPage, pageData.NumPages)

	logger.InfoContext(ctx, "successfully loaded profile info")
	h.renderer.Render(w, r, http.StatusOK, "profiles/show.gohtml", templateData)
}

// ProfileEditPageHandler displays the profile edit form.
func (h *Handlers) ProfileEditPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "ProfileEditPageHandler")

	profile, err := h.getProfileFromSession(r)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			logger.ErrorContext(ctx, "did not find profile in db", "profileID", profile.ID)
			status = http.StatusNotFound
		}

		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		h.Logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	templateData := NewProfilesTemplateData(r, w, "/profile", h.renderer)
	templateData.SetProfile(profile)
	h.renderer.Render(w, r, http.StatusOK, "profiles/edit.gohtml", templateData)
}

// ProfileEditHandler processes profile edit form submission.
func (h *Handlers) ProfileEditHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "ProfileEditHandler")

	profile, err := h.getProfileFromSession(r)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrNoRecord) {
			logger.ErrorContext(ctx, "did not find profile in db", slog.Any("error", err))
			status = http.StatusNotFound
		}

		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error loading your profile, please try logging in again")
		h.Logout(w, r)
		http.Redirect(w, r, "/login", status)
		return
	}

	req, err := parseEditProfileForm(r)
	if err != nil {
		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error editing your profile, please try again")
		templateData := NewProfilesTemplateData(r, w, "/profile", h.renderer)
		templateData.SetProfile(profile)
		h.renderer.Render(w, r, http.StatusBadRequest, "profiles/edit.gohtml", templateData)
		return
	}

	logger.InfoContext(ctx, "req to update profile", "req", req)

	err = profile.Update(ctx, logger, req)
	if err != nil {
		var editErr *identityaccess.ProfileEditValidationError

		h.flashMessager.SetErrorFlashMessage(w, r, "There was an error editing your profile, please try again")

		templateData := NewProfilesTemplateData(r, w, "/profile", h.renderer)
		templateData.SetProfile(profile)

		if errors.As(err, &editErr) {
			templateData.InitHasErrorFields()
			email, password, firstName, lastName := templateData.GetErrorFields()

			if editErr.EmailError != nil {
				**email = true
			}

			if editErr.PasswordError != nil {
				**password = true
			}

			if editErr.FirstNameError != nil {
				**firstName = true
			}

			if editErr.LastNameError != nil {
				**lastName = true
			}
		}

		h.renderer.Render(w, r, http.StatusBadRequest, "profiles/edit.gohtml", templateData)
		return

	}

	h.flashMessager.SetInfoFlashMessage(w, r, "Edited your profile!")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// GetPaginatedWatchHistoryHandler returns paginated watch history for the user's profile.
func (h *Handlers) GetPaginatedWatchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.With("handler", "GetPaginatedWatchHistoryHandler")
	logger.DebugContext(ctx, "getting paginated movies list")

	profileID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	page := r.URL.Query().Get("page")

	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	movieData, err := h.profileAggregatorService.GetWatchPaginatedHistory(ctx, logger, profileID, services.PageInfo{PageNum: pageNum})
	if err != nil {
		h.renderer.ServerError(w, r, err)
		return
	}

	if r.Header.Get("HX-Request") != "" {
		templateData := NewProfilesTemplateData(r, w, "/profile", h.renderer)
		templateData.SetWatchedMovies(movieData.WatchedMovies)
		templateData.SetPagination(pageNum, movieData.NumPages)
		h.renderer.RenderPartial(w, r, http.StatusOK, "profiles/partials/watch_list.gohtml", templateData)
		return

	}
}

// parseEditProfileForm parses the profile edit form data.
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

// getProfileFromSession retrieves the profile from the session.
func (h *Handlers) getProfileFromSession(r *http.Request) (*identityaccess.Profile, error) {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "identityaccess.http.Handlers.getProfileFromSession")
	defer span.End()
	profileID, err := h.sessionStore.GetProfileIDFromSession(ctx, r)
	if err != nil {
		return nil, err
	}

	profile, err := h.profileService.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	return profile, nil
}
