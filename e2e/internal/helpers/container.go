package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

	return dbCtr
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
			"DB_USERNAME":           "postgres",
			"DB_PASSWORD":           "postgres",
			"DB_MIGRATION_USER":     "postgres",
			"DB_MIGRATION_PASSWORD": "postgres",
			"DB_HOST":               dbIP,
			"DB_DATABASE_NAME":      "movieswithfriends",
			"TMDB_API_KEY":          tmdbKey,
			"SESSION_KEY":           sessionKey,
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
