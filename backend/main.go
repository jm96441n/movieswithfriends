package main

import (
	_ "embed"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/jm96441n/movieswithfriends/web"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("setting up postgres store")

	creds, err := store.NewCreds(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := store.NewPostgesStore(creds, os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE_NAME"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("completed postgres store setup")

	logger.Info("setting up session store")

	sessionStore := sessions.NewCookieStore(
		[]byte(os.Getenv("SESSION_KEY")),
	)

	sessionStore.Options = &sessions.Options{
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}
	logger.Info("completed session store setup")

	router := mux.NewRouter()

	web.SetupWebServer(logger, router, db, sessionStore)
	logger.Info("Listening on :8080...")

	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
