package identityaccess_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/testhelpers"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testContainerDB *postgres.PostgresContainer

func TestMain(m *testing.M) {
	ctx := context.Background()
	testContainerDB = &postgres.PostgresContainer{}

	err := testhelpers.SetupDBContainer(ctx, testContainerDB)

	defer func() {
		testContainerDB.Terminate(ctx)
	}()
	if err != nil {
		log.Printf("Failed to setup DB container: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func SetupConnPool(ctx context.Context, t *testing.T, schemaName string) *pgxpool.Pool {
	t.Helper()
	return testhelpers.SetupConnPool(ctx, t, schemaName, testContainerDB)
}
