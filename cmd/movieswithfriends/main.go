package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/services"
	iamstore "github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/migrations"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/ui"
	"github.com/jm96441n/movieswithfriends/web"
)

const length = 32

func main() {
	ctx := context.Background()
	var version string

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Error("Error reading build info")
		os.Exit(1)
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "vcs.revision" {
			slog.Info("Git version", slog.Any("vcs.revision", setting.Value))
			version = setting.Value
			break
		}
	}

	if version == "" {
		slog.Error("Could not find version in buildinfo")
	}

	collectorEndpoint := os.Getenv("COLLECTOR_ENDPOINT")
	if collectorEndpoint == "" {
		slog.Error("COLLECTOR_ENDPOINT is not set")
		os.Exit(1)
	}

	slog.Info("Got collector endpoint", slog.String("endpoint", collectorEndpoint))

	telemetry, err := metrics.NewTelemetry(ctx, metrics.Config{
		ServiceName:       "movieswithfriends",
		ServiceVersion:    version,
		Enabled:           true,
		CollectorEndpoint: collectorEndpoint,
	})
	if err != nil {
		slog.Error("Error building telemetry", slog.Any("err", err))
		os.Exit(1)
	}

	defer telemetry.Shutdown(ctx)

	logger := telemetry.Logger()
	logger.Info("successfully setup otel")

	logger.Info("running migrations")

	migrationUser := os.Getenv("DB_MIGRATION_USER")
	if migrationUser == "" {
		logger.Error("DB_MIGRATION_USER is not set")
		os.Exit(1)
	}

	migrationPassword := os.Getenv("DB_MIGRATION_PASSWORD")
	if migrationPassword == "" {
		logger.Error("DB_MIGRATION_PASSWORD is not set")
		os.Exit(1)
	}

	migrationCreds, err := newDBCreds(migrationUser, migrationPassword)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	err = runMigrations(logger, os.Getenv("DB_HOST"), os.Getenv("DB_DATABASE_NAME"), migrationCreds)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

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

	moviesRepo := partymgmtstore.NewMoviesRepository(connPool)

	assetDir, err := fs.Sub(ui.TemplateFS, "dist")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	loader, err := web.NewLoader(assetDir, "manifest.json")
	if err != nil {
		logger.Error("error loading manifest", slog.Any("err", err.Error()))
		os.Exit(1)
	}

	profileRepo := iamstore.NewProfileRepository(connPool)
	watcherRepo := partymgmtstore.NewWatcherRepository(connPool)
	partyRepo := partymgmtstore.NewPartyRepository(connPool)
	invitationsRepo := partymgmtstore.NewInvitationsRepository(connPool)

	partySvc := partymgmt.NewPartyService(logger, partyRepo)
	watcherSvc := partymgmt.NewWatcherService(watcherRepo)

	app := web.NewApplication(
		web.AppConfig{
			Telemetry:         telemetry,
			Logger:            logger,
			SessionStore:      sessionStore,
			MoviesService:     partymgmt.NewMovieService(tmdbClient, moviesRepo),
			MoviesRepository:  moviesRepo,
			PartyService:      partySvc,
			PartiesRepository: partyRepo,
			ProfilesService:   identityaccess.NewProfileService(profileRepo),
			WatcherService:    watcherSvc,
			Auth: &identityaccess.Authenticator{
				ProfileRepository: profileRepo,
			},
			ProfileAggregatorService: services.NewProfileAggregatorService(
				profileRepo,
				watcherRepo,
				partySvc,
				watcherSvc,
			),
			InvitationsService: partymgmt.NewInvitationsService(invitationsRepo),
			AssetLoader:        loader,
		},
	)

	addr := os.Getenv("ADDR")
	if addr == "" {
		logger.Info("ADDR is not set, defaulting to :4000")
		addr = ":4000"
	}

	server := http.Server{
		Addr: addr,
		Handler: otelhttp.NewHandler(app.Routes(), "movieswithfriends",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
			otelhttp.WithMeterProvider(telemetry.MeterProvider()),
			otelhttp.WithTracerProvider(telemetry.TracerProvider()),
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
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

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func runMigrations(logger *slog.Logger, host, dbname string, creds DBCreds) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	goose.SetBaseFS(migrations.MigrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", creds.Username, creds.Password, host, dbname)

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return err
	}

	logger.Info("opened db, starting ping")

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}
	logger.Info("pinged db, running goose up")

	err = goose.Up(db, "migrations")
	if err != nil {
		return err
	}

	logger.Info("ran goose up, closing db")
	err = db.Close()
	if err != nil {
		return err
	}

	logger.Info("closed db for migrations")

	return nil
}

var (
	ErrMissingDBUsername     = errors.New("DB_USERNAME env var is missing")
	ErrMissingDBPassword     = errors.New("DB_PASSWORD env var is missing")
	ErrMissingDBHost         = errors.New("DB_HOST env var is missing")
	ErrMissingDBDatabaseName = errors.New("DB_DATABASE_NAME env var is missing")
)
