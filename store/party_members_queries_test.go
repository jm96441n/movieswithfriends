package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPartyMembersQueries(t *testing.T) {
	ctx := context.Background()

	partyMemberDBContainer := setupDBContainer(ctx, t)

	err := partyMemberDBContainer.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]func(t *testing.T){
		"testCreatePartyMember": testCreatePartyMember(ctx, partyMemberDBContainer),
	}

	for name, tc := range testCases {
		t.Run(name, tc)
	}
}

func testCreatePartyMember(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(func() {
			testConn.Close(ctx)
			db.Close()
			err := ctr.Restore(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})

		idMember, idParty := seedCreatePartyMemberBackground(t, testConn)

		err := db.CreatePartyMember(context.Background(), idMember, idParty)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func seedCreatePartyMemberBackground(t *testing.T, conn *pgx.Conn) (int, int) {
	t.Helper()
	var (
		idMember int
		idParty  int
	)
	err := conn.QueryRow(context.Background(), "insert into parties (name, short_id) values($1, $2) returning id_party", "test-party", "abcdef").Scan(&idParty)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(context.Background(), "insert into profiles (first_name, last_name) values($1, $2) returning id_profile", "tom", "bomba").Scan(&idMember)
	if err != nil {
		t.Fatal(err)
	}
	return idMember, idParty
}

func setupDBContainer(ctx context.Context, t *testing.T) *postgres.PostgresContainer {
	t.Helper()
	partyMemberDBContainer, err := postgres.Run(
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

	testcontainers.CleanupContainer(t, partyMemberDBContainer)
	if err != nil {
		t.Fatal(err)
	}

	goose.SetBaseFS(os.DirFS("../migrations"))

	if err = goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	connString, err := partyMemberDBContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	err = goose.Up(db, ".")
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	return partyMemberDBContainer
}

func setupDB(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer) (*store.PGStore, *pgx.Conn) {
	t.Helper()

	host, err := ctr.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := ctr.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.NewPostgesStore(
		store.Creds{
			Username: "postgres",
			Password: "postgres",
		},
		fmt.Sprintf("%s:%s", host, port.Port()),
		"movieswithfriends",
		&slog.Logger{},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping(ctx)
	if err != nil {
		t.Fatal(err)
	}

	connString, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	testConn, err := pgx.Connect(ctx, connString)
	if err != nil {
		t.Fatal(err)
	}

	return db, testConn
}
