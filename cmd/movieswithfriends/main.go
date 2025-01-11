package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	iamstore "github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/jm96441n/movieswithfriends/ui"
	"github.com/jm96441n/movieswithfriends/web"

	"github.com/gorilla/sessions"
	"github.com/honeycombio/otel-config-go/otelconfig"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
)

const length = 32

func main() {
	ctx := context.Background()

	// Create the OTLP log exporter that sends logs to configured destination
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		panic("failed to initialize exporter")
	}

	// Create the logger provider
	lp := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(logExporter),
		),
	)

	// Ensure the logger is shutdown before exiting so all pending logs are exported
	defer lp.Shutdown(ctx)

	// Set the logger provider globally
	global.SetLoggerProvider(lp)

	// Instantiate a new slog logger
	logger := slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			otelslog.NewHandler("movieswithfriends"),
		))

	// use otelconfig to set up OpenTelemetry SDK
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry()
	if err != nil {
		logger.Error("error setting up OTel SDK", slog.Any("err", err))
		os.Exit(1)
	}

	defer otelShutdown()

	logger.Info("successfully setup otel")

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

	creds, err := newDBCreds(dbUser, dbPassword)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	connPool, err := createConnPool(os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE_NAME"), creds)
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

	tmdbClient, err := partymgmt.NewTMDBClient("https://api.themoviedb.org/3", tmdbApiKey, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionKey := make([]byte, length)
	sessionKeyVar := os.Getenv("SESSION_KEY")
	if sessionKeyVar == "" {
		logger.Error("SESSION_KEY is not set")

		if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
			logger.Error("could not generate secure key", slog.Any("err", err))
			os.Exit(1)
		}
	} else {
		sessionKey = []byte(sessionKeyVar)
	}

	sessionStore := sessions.NewCookieStore(sessionKey)

	db, err := store.NewPostgesStore(store.Creds(creds), os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE_NAME"), logger)
	if err != nil {
		logger.Error("could not generate secure key", slog.Any("err", err))
		os.Exit(1)
	}

	app := web.Application{
		TemplateCache:     tmplCache,
		Logger:            logger,
		SessionStore:      sessionStore,
		MoviesService:     partymgmt.NewMovieService(tmdbClient, db),
		MoviesRepository:  db,
		PartyService:      &partymgmt.PartyService{DB: db, MoviesRepository: db, Logger: logger},
		PartiesRepository: db,
		ProfilesService: identityaccess.NewProfileService(
			iamstore.NewProfileRepository(connPool),
		),
		MemberService: partymgmt.NewMemberService(db),
		Auth: &identityaccess.Authenticator{
			Logger:            logger,
			ProfileRepository: iamstore.NewProfileRepository(connPool),
		},
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		logger.Info("ADDR is not set, defaulting to :4000")
		addr = ":4000"
	}

	server := http.Server{
		Addr: addr,
		Handler: otelhttp.NewHandler(app.Routes(), "movieswithfriends",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", slog.String("addr", addr))

	err = server.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

type DBCreds struct {
	Username string
	Password string
}

func newDBCreds(username, pw string) (DBCreds, error) {
	if username == "" {
		return DBCreds{}, ErrMissingDBUsername
	}

	if pw == "" {
		return DBCreds{}, ErrMissingDBPassword
	}
	return DBCreds{
		Username: username,
		Password: url.QueryEscape(pw),
	}, nil
}

func createConnPool(host string, dbname string, creds DBCreds) (*pgxpool.Pool, error) {
	if host == "" {
		return nil, ErrMissingDBHost
	}

	if dbname == "" {
		return nil, ErrMissingDBDatabaseName
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", creds.Username, creds.Password, host, dbname)
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	db.Ping(ctx)
	return db, nil
}

var (
	ErrMissingDBUsername     = errors.New("DB_USERNAME env var is missing")
	ErrMissingDBPassword     = errors.New("DB_PASSWORD env var is missing")
	ErrMissingDBHost         = errors.New("DB_HOST env var is missing")
	ErrMissingDBDatabaseName = errors.New("DB_DATABASE_NAME env var is missing")
)
