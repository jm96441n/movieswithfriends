package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupDBContainer(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	t.Helper()
	dbCtr, err := postgres.Run(
		ctx,
		"postgres:bullseye",
		postgres.WithDatabase("movieswithfriends"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second)),
	)

	testcontainers.CleanupContainer(t, dbCtr)
	if err != nil {
		t.Fatalf("failed to run postgres container: %v", err)
	}

	goose.SetBaseFS(os.DirFS("../migrations"))

	if err = goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}
	connString, err := dbCtr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	err = goose.Up(db, ".")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = dbCtr.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return dbCtr
}

func SetupDBConn(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer) *pgx.Conn {
	t.Helper()

	connString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	testConn, err := pgx.Connect(ctx, connString)
	if err != nil {
		t.Fatal(err)
	}

	return testConn
}

func CleanupAndResetDB(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer, testConn *pgx.Conn) func() {
	return func() {
		testConn.Close(ctx)
		err := ctr.Restore(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

type logConsumer struct{}

func (g *logConsumer) Accept(l testcontainers.Log) {
	fmt.Println(string(l.Content))
}

func SetupAppContainer(ctx context.Context, t *testing.T, pgCtr *postgres.PostgresContainer) testcontainers.Container {
	dbIP, err := pgCtr.ContainerIP(ctx)
	if err != nil {
		t.Fatalf("Failed to get container IP: %v", err)
	}

	tmdbKey := os.Getenv("TMDB_API_KEY")
	sessionKey := os.Getenv("SESSION_KEY")

	req := testcontainers.ContainerRequest{
		Image:        "movieswithfriends:test",
		ExposedPorts: []string{"4000/tcp"},
		Networks:     []string{"bridge", "test"},
		WaitingFor:   wait.ForHTTP("/health"),
		Env: map[string]string{
			"DB_USERNAME":      "postgres",
			"DB_PASSWORD":      "postgres",
			"DB_HOST":          dbIP,
			"DB_DATABASE_NAME": "movieswithfriends",
			"TMDB_API_KEY":     tmdbKey,
			"SESSION_KEY":      sessionKey,
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&logConsumer{}},
		},
	}

	appCtr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to bring up app container: %v", err)
	}

	return appCtr
}
