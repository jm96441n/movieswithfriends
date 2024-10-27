package main

import (
	"crypto/rand"
	"crypto/tls"
	_ "embed"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/jm96441n/movieswithfriends/ui"
	"github.com/jm96441n/movieswithfriends/web"
)

const length = 32

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("setting up postgres store")
	dbUser := os.Getenv("DB_USERNAME")
	if dbUser == "" {
		logger.Error("DB_USERNAME is not set")
		os.Exit(1)
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		logger.Error("DB_PASSWORD is not set")
		os.Exit(1)
	}

	creds, err := store.NewCreds(dbUser, dbPassword)
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

	tmdbClient, err := partymgmt.NewTMDBClient("https://api.themoviedb.org/3", tmdbApiKey)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionKey := make([]byte, length)
	sessionKeyVar := os.Getenv("SESSION_KEY")
	if sessionKeyVar == "" {
		logger.Error("SESSION_KEY is not set")

		if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
			log.Fatalf("could not generate secure key: %v", err)
			os.Exit(1)
		}
	} else {
		sessionKey = []byte(sessionKeyVar)
	}

	sessionStore := sessions.NewCookieStore([]byte(sessionKey))

	app := web.Application{
		TemplateCache:     tmplCache,
		Logger:            logger,
		SessionStore:      sessionStore,
		MoviesService:     partymgmt.NewMovieService(tmdbClient, logger, db),
		MoviesRepository:  db,
		PartyService:      &partymgmt.PartyService{DB: db},
		PartiesRepository: db,
		ProfilesService:   db,
		Auth: &identityaccess.Authenticator{
			Logger:            logger,
			AccountRepository: db,
		},
		AccountRepository: db,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		logger.Info("ADDR is not set, defaulting to :4000")
		addr = ":4000"
	}

	tlsCert := os.Getenv("TLS_CERT_LOCATION")
	if tlsCert == "" {
		logger.Error("TLS_CERT_LOCATION is not set")
		os.Exit(1)
	}

	tlsKey := os.Getenv("TLS_KEY_LOCATION")
	if tlsKey == "" {
		logger.Error("TLS_KEY_LOCATION is not set")
		os.Exit(1)
	}
	// binding.pry
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

	err = server.ListenAndServeTLS(tlsCert, tlsKey)
	logger.Error(err.Error())
	os.Exit(1)
}
