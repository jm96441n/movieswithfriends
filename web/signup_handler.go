package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
)

type SignupResponse struct {
	Message string
}

func (a *Application) SignUpShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, w, "/signup")
	a.render(w, r, http.StatusOK, "signup/show.gohtml", data)
}

func (a *Application) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*500)
	r.WithContext(ctx)

	defer cancel()
	defer r.Body.Close()

	req, err := parseSignUpForm(r)
	if err != nil {
		a.Logger.Error("error parsing signup form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	_, err = a.Auth.CreateAccount(ctx, req)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Debug("seeting flash message")
	a.setFlashMessage(r, w, "Successfully signed up! Please log in.")

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
		PartyID:   r.FormValue("partyID"),
	}, nil
}
