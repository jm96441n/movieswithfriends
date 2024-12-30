package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func Setup(ctx context.Context, t *testing.T, testConn *pgxpool.Pool, page playwright.Page) {
	t.Helper()
	t.Cleanup(func() {
		t.Logf("Cleaning up from test")
		cleanupAndResetDB(ctx, t, testConn)
		err := page.Context().ClearCookies()
		if err != nil {
			t.Fatalf("could not clear cookies: %v", err)
		}
	})
}

func SetupDBConnPool(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer) *pgxpool.Pool {
	t.Helper()

	connString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	connPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		connPool.Close()
	})

	return connPool
}

func cleanupAndResetDB(ctx context.Context, t *testing.T, testConn *pgxpool.Pool) {
	t.Helper()
	// this truncats all tables, using the snapshot restore functionality from testcontainers was causing connection drop errors between tests
	_, err := testConn.Exec(ctx, `DO $$ 
DECLARE 
    r RECORD;
BEGIN
    -- Disable foreign key checks temporarily
    SET CONSTRAINTS ALL DEFERRED;
    
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;`)
	ok(t, err)
}
