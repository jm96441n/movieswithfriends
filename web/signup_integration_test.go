// go:build integration

package web_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/playwright-community/playwright-go"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	m.Run()
}

func TestSignup(t *testing.T) {
	ctx := context.Background()
	ctr := setupDBContainer(ctx, t)

	db, testConn := setupDB(ctx, t, ctr)
	t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto("http://localhost:4000/signup"); err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	fillInField(t, "First Name", "Buddy", page)
	fillInField(t, "Last Name", "TheElf", page)
	fillInField(t, "Email", "buddy3@santa.com", page)
	fillInField(t, "Password", "1Password", page)

	err = page.GetByText("Create Account").Click()
	if err != nil {
		t.Fatalf("could not click create account button: %v", err)
	}

	flashMsg := page.GetByText("Successfully signed up! Please log in.")
	if flashMsg == nil {
		t.Fatalf("expected success message, got nil")
	}

	curURL := page.URL()
	if !strings.Contains(curURL, "/login") {
		t.Fatalf("expected to be on login page, got %s", curURL)
	}
}

func fillInField(t *testing.T, label, value string, page playwright.Page) {
	t.Helper()
	field := page.GetByLabel(label)

	err := field.Fill(value)
	if err != nil {
		t.Fatalf("could not fill %s: %v", label, err)
	}
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

	err = partyMemberDBContainer.Snapshot(context.Background())
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

func cleanupAndResetDB(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer, testConn *pgx.Conn, db *store.PGStore) func() {
	return func() {
		testConn.Close(ctx)
		db.Close()
		err := ctr.Restore(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}
