package main

import (
	"crypto/tls"
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
	"github.com/jm96441n/movieswithfriends/ui"
	"github.com/jm96441n/movieswithfriends/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("setting up postgres store")

	creds, err := store.NewCreds(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := store.NewPostgesStore(creds, os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE_NAME"), logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("completed postgres store setup")

	tmdbApiKey := os.Getenv("TMDB_API_KEY")
	if tmdbApiKey == "" {
		logger.Error("TMDB_API_KEY is not set")
		os.Exit(1)
	}

	tmplCache, err := web.NewTemplateCache(ui.TemplateFS)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	tmdbClient := web.NewTMDBClient("https://api.themoviedb.org/3", tmdbApiKey)

	app := web.Application{
		TemplateCache:   tmplCache,
		Logger:          logger,
		TMDBClient:      tmdbClient,
		MoviesService:   db,
		PartiesService:  db,
		ProfilesService: db,
		AccountService:  db,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	addr := ":4000"

	server := http.Server{
		Addr:         addr,
		Handler:      app.Routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", slog.String("addr", addr))

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}
