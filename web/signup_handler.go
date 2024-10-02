package web

import (
	"context"
	"errors"
	"fmt"
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

func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/login")
	a.render(w, r, http.StatusOK, "login/show.gohtml", data)
}

func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.Logger.Error("error parsing form", slog.Any("error", err))
		a.serverError(w, r, err)
		return
	}

	account, err := a.AccountService.FindAccountByEmail(r.Context(), r.FormValue("email"))
	if err != nil {
		a.Logger.Error("error finding account by email", slog.Any("error", err), slog.String("email", r.FormValue("email")))
		data := a.NewTemplateData(r, "/signup")
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
		return
	}

	err = bcrypt.CompareHashAndPassword(account.Password, []byte(r.PostForm.Get("password")))
	if err != nil {
		data := a.NewTemplateData(r, "/signup")
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			a.Logger.Error("error comparing password", slog.Any("error", err))
			a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
			return
		}
		a.Logger.Error("error comparing password", slog.Any("error", err))
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", data)
		return
	}
	session, err := a.SessionStore.Get(r, sessionName)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	session.Values["accountID"] = account.ID
	session.Values["profileID"] = account.Profile.ID

	err = session.Save(r, w)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/profiles/%d", account.Profile.ID), http.StatusSeeOther)
}
