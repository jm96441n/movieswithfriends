package helpers

import (
	"context"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func Setup(ctx context.Context, t *testing.T, testConn *pgxpool.Pool, page playwright.Page) {
	t.Helper()
	t.Cleanup(func() {
		t.Logf("Cleaning up from test")
		cleanupAndResetDB(ctx, t, testConn)
		Ok(t, page.Context().ClearCookies(), "faild to clear browser cookies")
	})
}

func SetupSuite(ctx context.Context, t *testing.T) (*pgxpool.Pool, playwright.Page, nat.Port) {
	t.Helper()
	dbCtr := SetupDBContainer(ctx, t)
	appCtr := SetupAppContainer(ctx, t, dbCtr)

	pw, err := playwright.Run()
	Ok(t, err, "could not start playwright")

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	Ok(t, err, "could not launch browser")

	port, err := appCtr.MappedPort(ctx, "4000")
	Ok(t, err, "could not get port mapping")

	page, err := browser.NewPage()
	Ok(t, err, "could not create page")

	connPool := SetupDBConnPool(ctx, t, dbCtr)

	return connPool, page, port
}

func SetupDBConnPool(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer) *pgxpool.Pool {
	t.Helper()

	connString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	Ok(t, err, "failed to get connection string for container")

	connPool, err := pgxpool.New(ctx, connString)
	Ok(t, err, "failed to create connection pool")

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
	Ok(t, err, "failed to truncate test db")
}
