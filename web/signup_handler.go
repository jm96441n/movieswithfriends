package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess"
)

type SignupResponse struct {
	Message string
}

func (a *Application) SignUpShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/signup")
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

	fmt.Println("req: ", req)

	_, err = a.Auth.CreateAccount(ctx, req)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Info("successfully signed up user", "userName", req.FirstName, "userEmail", req.Email)
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
