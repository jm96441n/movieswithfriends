package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	cmd := exec.Command("docker", "build", "--target=prod", "-t", "movieswithfriends:test", ".")
	cmd.Dir = ".."

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		log.Fatalf("could not build docker image: %v", err)
	}

	m.Run()
}

func TestSignup(t *testing.T) {
	ctx := context.Background()
	dbCtr := setupDBContainer(ctx, t)
	appCtr := setupAppContainer(ctx, t, dbCtr)

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}

	port, err := appCtr.MappedPort(ctx, "4000")
	if err != nil {
		t.Fatalf("failed to get port mapping: %v", err)
	}

	tests := map[string]func(*testing.T){
		"testSignupIsSuccessful":             testSignupIsSuccessful(browser, port.Port()),
		"testSignupFailsIfEmailIsInUse":      testSignupFailsIfEmailIsInUse(dbCtr, browser, port.Port()),
		"testSignupFailsWithFormValidations": testSignupFailsWithFormValidations(browser, port.Port()),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testSignupIsSuccessful(browser playwright.Browser, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		page, err := browser.NewPage()
		if err != nil {
			t.Fatalf("could not create page: %v", err)
		}

		if _, err = page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort)); err != nil {
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

		fillInField(t, "Email Address", "buddy3@santa.com", page)
		fillInField(t, "Password", "1Password", page)

		err = page.Locator("button").GetByText("Sign In").Click()
		if err != nil {
			t.Fatalf("could not click sign in button: %v", err)
		}

		curURL = page.URL()
		if !strings.Contains(curURL, "/profile") {
			t.Fatalf("expected to be on profile page, got %s", curURL)
		}
	}
}

func testSignupFailsIfEmailIsInUse(dbCtr *postgres.PostgresContainer, browser playwright.Browser, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()
		testConn := setupDBConn(ctx, t, dbCtr)
		t.Cleanup(cleanupAndResetDB(ctx, t, dbCtr, testConn))

		seedAccount(ctx, t, testConn, "buddy@santa.com", "anotherpassword")

		page, err := browser.NewPage()
		if err != nil {
			t.Fatalf("could not create page: %v", err)
		}

		if _, err = page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort)); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		fillInField(t, "First Name", "Buddy", page)
		fillInField(t, "Last Name", "TheElf", page)
		fillInField(t, "Email", "buddy@santa.com", page)
		fillInField(t, "Password", "1Password", page)

		err = page.GetByText("Create Account").Click()
		if err != nil {
			t.Fatalf("could not click create account button: %v", err)
		}

		locatorChecker := playwright.NewPlaywrightAssertions()

		flashMsg := page.GetByText("An account exists with this email. Try logging in or resetting your password.")
		flashChecker := locatorChecker.Locator(flashMsg)
		if err := flashChecker.Not().ToBeEmpty(); err != nil {
			t.Fatal("expected error message in flash, got nothing")
		}

		regex := regexp.MustCompile(`.*alert-danger.*`)

		if err := flashChecker.ToHaveClass(regex); err != nil {
			s, err := flashMsg.GetAttribute("class")
			if err != nil {
				t.Fatalf("failed to get class attribute: %v", err)
			}
			t.Fatalf("expected flash message to be have class \"alert-danger\", it was %s", s)
		}

		curURL := page.URL()
		if !strings.Contains(curURL, "/signup") {
			t.Fatalf("expected to be on signup page, got %s", curURL)
		}
	}
}

func testSignupFailsWithFormValidations(browser playwright.Browser, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		page, err := browser.NewPage()
		if err != nil {
			t.Fatalf("could not create page: %v", err)
		}

		if _, err = page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort)); err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		err = page.GetByText("Create Account").Click()
		if err != nil {
			t.Fatalf("could not click create account button: %v", err)
		}

		errMsgs := page.Locator(".invalid-feedback")
		if errMsgs == nil {
			t.Fatalf("could not get error messages")
		}

		texts, err := errMsgs.AllInnerTexts()
		if err != nil {
			t.Fatalf("could not get error messages text: %v", err)
		}

		for _, text := range texts {
			fmt.Println(text)
		}
	}
}

func seedAccount(ctx context.Context, t *testing.T, conn *pgx.Conn, email, password string) {
	t.Helper()
	_, err := conn.Exec(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2)", email, password)
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("failed to run postgres container: %v", err)
	}

	goose.SetBaseFS(os.DirFS("../migrations"))

	if err = goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}
	connString, err := partyMemberDBContainer.ConnectionString(ctx, "sslmode=disable")
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

	err = partyMemberDBContainer.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	return partyMemberDBContainer
}

func setupDBConn(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer) *pgx.Conn {
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

func cleanupAndResetDB(ctx context.Context, t *testing.T, ctr *postgres.PostgresContainer, testConn *pgx.Conn) func() {
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

func setupAppContainer(ctx context.Context, t *testing.T, pgCtr *postgres.PostgresContainer) testcontainers.Container {
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
