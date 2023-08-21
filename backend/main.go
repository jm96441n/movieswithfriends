package main

import (
	"embed"
	_ "embed"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/jm96441n/movieswithfriends/web"
	"golang.org/x/exp/slog"
)

//go:embed all:templates

var templateFS embed.FS

func main() {
	tmpls := web.BuildTemplates(templateFS)
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

	router := mux.NewRouter()

	web.SetupWebServer(logger, router, db, tmpls)
	logger.Info("Listening on :8080...")

	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
