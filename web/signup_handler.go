package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SignupReq struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	PartyID   string `json:"partyID"`
}

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	_, err = a.AccountService.CreateAccount(ctx, req.Email, req.FirstName, req.LastName, hashedPassword)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Info("successfully signed up user", "userName", req.FirstName, "userEmail", req.Email)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func parseSignUpForm(r *http.Request) (SignupReq, error) {
	err := r.ParseForm()
	if err != nil {
		return SignupReq{}, err
	}
	return SignupReq{
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
		FirstName: r.FormValue("firstName"),
		LastName:  r.FormValue("lastName"),
		PartyID:   r.FormValue("partyID"),
	}, nil
}
