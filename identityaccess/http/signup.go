package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
)

// SignUpShowHandler displays the signup form.
func (h *Handlers) SignUpShowHandler(w http.ResponseWriter, r *http.Request) {
	data := NewSignupTemplateData(r, w, "/signup", h.renderer)
	h.renderer.Render(w, r, http.StatusOK, "signup/show.gohtml", data)
}

// SignUpHandler processes the signup form submission.
func (h *Handlers) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*500)
	r = r.WithContext(ctx)
	logger := h.logger.With("handler", "SignUpHandler")

	defer cancel()
	defer r.Body.Close()

	req, err := parseSignUpForm(r)
	if err != nil {
		logger.ErrorContext(ctx, "error parsing signup form", slog.Any("error", err))
		h.renderer.ServerError(w, r, err)
		return
	}

	_, err = h.profileService.CreateProfile(ctx, logger, req)
	if err != nil {
		if errors.Is(err, identityaccess.ErrAccountExists) {
			h.flashMessager.SetErrorFlashMessage(w, r, "An account exists with this email. Try logging in or resetting your password.")
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}

		var signupErr *identityaccess.SignupValidationError

		if errors.As(err, &signupErr) {
			// Get signup-specific template data
			data := NewSignupTemplateData(r, w, "/signup", h.renderer)
			data.InitHasErrorFields()

			// Get pointers to error fields
			email, password, firstName, lastName := data.GetErrorFields()

			if signupErr.EmailError != nil {
				**email = true
			}
			if signupErr.PasswordError != nil {
				**password = true
			}
			if signupErr.FirstNameError != nil {
				**firstName = true
			}
			if signupErr.LastNameError != nil {
				**lastName = true
			}

			h.renderer.Render(w, r, http.StatusBadRequest, "signup/show.gohtml", data)
			return
		}

		h.renderer.ServerError(w, r, err)
		return
	}

	logger.DebugContext(ctx, "seeting flash message")
	h.flashMessager.SetInfoFlashMessage(w, r, "Successfully signed up! Please log in.")

	h.telemetry.IncreaseUserRegisteredCounter(ctx, logger)

	logger.DebugContext(ctx, "successfully signed up user", "userName", req.FirstName, "userEmail", req.Email)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// parseSignUpForm extracts signup data from the HTTP request form.
func parseSignUpForm(r *http.Request) (identityaccess.SignupReq, error) {
	err := r.ParseForm()
	if err != nil {
		return identityaccess.SignupReq{}, err
	}

	return identityaccess.SignupReq{
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
	}, nil
}
