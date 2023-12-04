package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/exp/slog"
)

type profileFinder interface {
	GetProfile(context.Context, string) (store.Profile, error)
}

func ProfileShowHandler(logger *slog.Logger, db profileFinder, sessionStore sessions.Store) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()
		logger.Info("Received request for profile show")
		session, err := sessionStore.Get(r, sessionName)
		if err != nil {
			logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user := getUser(session)
		logger.Info(user.Login)
		profile, err := db.GetProfile(ctx, user.Login)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}
		logger.Info(fmt.Sprintf("%+v", profile))

		body, err := json.Marshal(profile)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write(body)
		return
	})
}

func getUser(s *sessions.Session) User {
	val := s.Values["user"]
	user := User{}
	user, ok := val.(User)
	if !ok {
		return User{Authenticated: false}
	}
	return user
}
