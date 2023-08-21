package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

type profileFinder interface {
	GetProfile(context.Context, int) (store.Profile, error)
}

func ProfileShowHandler(logger *slog.Logger, db profileFinder, tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()
		logger.Info("Received request for profile show")
		profile, err := db.GetProfile(ctx, 1)
		pageData := map[string]string{"Name": profile.Name}
		err = tmpl.ExecuteTemplate(w, "show.gohtml", pageData)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
		}
		return
	})
}

func SignUpShowHandler(logger *slog.Logger, tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageData := make(map[string]string, 0)
		err := tmpl.ExecuteTemplate(w, "signup.gohtml", pageData)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}
		return
	})
}

func LoginShowHandler(logger *slog.Logger, tmpl *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageData := make(map[string]string, 0)
		err := tmpl.ExecuteTemplate(w, "login.gohtml", pageData)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}
		return
	})
}

type accountCreator interface {
	CreateAccount(context.Context, string, string, []byte) error
}

type SignupReq struct {
	Login    string
	Password string
	Name     string
	PartyID  string
}

type SignupResponse struct {
	Message string
}

func SignUpHandler(logger *slog.Logger, db accountCreator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		req := SignupReq{}

		err = json.Unmarshal(body, &req)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		err = db.CreateAccount(ctx, req.Name, req.Login, hashedPassword)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		resp := SignupResponse{Message: fmt.Sprintf("Successfully signed up user %s with login %s", req.Name, req.Login)}

		respBody, err := json.Marshal(resp)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		_, err = w.Write(respBody)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}
	})
}

type accountFinder interface {
	GetAccountByLogin(context.Context, string) (store.Account, error)
}

func LoginHandler(logger *slog.Logger, db accountCreator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		return
	})
}
