package web

import (
	"context"
	"html/template"
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

func SignUpHandler(logger *slog.Logger, db accountCreator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		r.ParseForm()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		err = db.CreateAccount(ctx, r.FormValue("name"), r.FormValue("login"), hashedPassword)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)

		return
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
