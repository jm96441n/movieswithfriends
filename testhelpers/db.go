package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var containerOnce = &sync.Once{}

func SetupConnPool(ctx context.Context, t *testing.T, schemaName string, testContainerDB *postgres.PostgresContainer) *pgxpool.Pool {
	t.Helper()

	connString, err := testContainerDB.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	connString, err = testContainerDB.ConnectionString(ctx, "sslmode=disable", fmt.Sprintf("search_path=%s", schemaName))
	if err != nil {
		t.Fatal(err)
	}

	err = runMigrations(connString)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatal(err)
	}

	return pool
}

func SetupDBContainer(ctx context.Context, ctr *postgres.PostgresContainer) error {
	testCtr, err := postgres.Run(
		ctx,
		"postgres:bullseye",
		postgres.WithDatabase("movieswithfriends"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	*ctr = *testCtr
	if err != nil {
		return err
	}
	return nil
}

func runMigrations(connString string) error {
	goose.SetBaseFS(os.DirFS("../../migrations"))

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return err
	}

	err = goose.Up(db, ".")
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}

func CleanupAndResetDB(ctx context.Context, t *testing.T, testConn *pgxpool.Pool, schemaName string) {
	t.Helper()
	t.Log("Cleaning up and resetting db")
	// this truncates all tables, using the snapshot restore functionality from testcontainers was causing connection drop errors between tests
	_, err := testConn.Exec(ctx, fmt.Sprintf(`DO $$ 
DECLARE 
    r RECORD;
BEGIN
    -- Disable foreign key checks temporarily
    SET CONSTRAINTS ALL DEFERRED;
    
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = '%s') LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;`, schemaName))
	Ok(t, err, "failed to truncate test db")
}
