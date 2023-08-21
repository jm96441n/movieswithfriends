package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
)

type accountFinder interface {
	GetAccountByLogin(context.Context, string) (store.Account, error)
}

type LoginReq struct {
	Login    string
	Password string
}

type LoginResp struct {
	Message string
}

type User struct {
	Login         string
	Authenticated bool
}

func LoginHandler(logger *slog.Logger, db accountFinder, sessionStore sessions.Store) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		session, err := sessionStore.Get(r, "moviescookie")
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req := LoginReq{}

		err = json.Unmarshal(body, &req)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		account, err := db.GetAccountByLogin(ctx, req.Login)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authenticated := authenticate(account, req)

		if !authenticated {
			logger.Error("username/password incorrect")
			http.Error(w, "username/password incorrect", http.StatusUnauthorized)
			return
		}

		userCookie := User{Login: req.Login, Authenticated: true}

		session.Values["user"] = userCookie

		err = session.Save(r, w)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := LoginResp{Message: fmt.Sprintf("Successfully logged in user %s", req.Login)}

		respBody, err := json.Marshal(resp)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(respBody)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func authenticate(account store.Account, req LoginReq) bool {
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)) == nil {
		return true
	}

	return false
}
