package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SignupReq struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	PartyID   string
}

type SignupResponse struct {
	Message string
}

func (a *Application) SignUpShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/signup")
	a.render(w, r, http.StatusOK, "signup/show.gohtml", data)
}

func (a *Application) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	req := SignupReq{}

	err = json.Unmarshal(body, &req)
	if err != nil {
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
	http.Redirect(w, r, "/profiles/1", http.StatusSeeOther)
}

func parseSignUpForm(r *http.Request) (SignupReq, error) {
	err := r.ParseForm()
	if err != nil {
		return SignupReq{}, err
	}
	return SignupReq{
		Email:     r.PostForm.Get("email"),
		Password:  r.PostForm.Get("password"),
		FirstName: r.PostForm.Get("firstName"),
		LastName:  r.PostForm.Get("lastName"),
		PartyID:   r.PostForm.Get("partyID"),
	}, nil
}

func (a *Application) LoginShowHandler(w http.ResponseWriter, r *http.Request) {
	data := a.NewTemplateData(r, "/login")
	a.render(w, r, http.StatusOK, "login/show.gohtml", data)
}

func (a *Application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	account, err := a.AccountService.FindAccountByEmail(r.Context(), r.PostForm.Get("email"))
	if err != nil {
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", nil)
		return
	}

	err = bcrypt.CompareHashAndPassword(account.Password, []byte(r.PostForm.Get("password")))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", nil)
			return
		}
		a.render(w, r, http.StatusUnauthorized, "login/show.gohtml", nil)
		return
	}
	http.Redirect(w, r, "/profiles/1", http.StatusSeeOther)
}
