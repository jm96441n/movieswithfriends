package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
)

type SignupResponse struct {
	Message string
}

func (a *Application) SignUpShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewSignupTemplateData(r, w, "/signup")
	a.render(w, r, http.StatusOK, "signup/show.gohtml", data)
}

func (a *Application) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*500)
	r = r.WithContext(ctx)
	logger := a.Logger.With("handler", "SignUpHandler")

	defer cancel()
	defer r.Body.Close()

	req, err := parseSignUpForm(r)
	if err != nil {
		a.Logger.Error("error parsing signup form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	_, err = a.ProfilesService.CreateProfile(ctx, logger, req)
	if err != nil {
		if errors.Is(err, identityaccess.ErrAccountExists) {
			a.setErrorFlashMessage(w, r, "An account exists with this email. Try logging in or resetting your password.")
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}

		var signupErr *identityaccess.SignupValidationError

		if errors.As(err, &signupErr) {
			data := a.NewSignupTemplateData(r, w, "/signup")
			data.InitHasErrorFields()
			if signupErr.EmailError != nil {
				*data.HasEmailError = true
			}

			if signupErr.PasswordError != nil {
				*data.HasPasswordError = true
			}

			if signupErr.FirstNameError != nil {
				*data.HasFirstNameError = true
			}

			if signupErr.LastNameError != nil {
				*data.HasLastNameError = true
			}

			a.render(w, r, http.StatusBadRequest, "signup/show.gohtml", data)
			return
		}

		a.serverError(w, r, err)
		return
	}

	a.Logger.Debug("seeting flash message")
	a.setInfoFlashMessage(w, r, "Successfully signed up! Please log in.")

	a.Telemetry.IncreaseUserRegisteredCounter(ctx, logger)

	a.Logger.Debug("successfully signed up user", "userName", req.FirstName, "userEmail", req.Email)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

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
