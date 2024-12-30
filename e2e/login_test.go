package e2e

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbCtr := helpers.SetupDBContainer(ctx, t)
	appCtr := helpers.SetupAppContainer(ctx, t, dbCtr)

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

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	connPool := helpers.SetupDBConnPool(ctx, t, dbCtr)

	tests := map[string]func(t *testing.T){
		"testLoginIsSuccessful":                           testLoginIsSuccessful(ctx, connPool, page, port.Port()),
		"testLoginFailsWhenUsernameOrPasswordIsIncorrect": testLoginFailsWhenUsernameOrPasswordIsIncorrect(ctx, connPool, page, port.Port()),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testLoginIsSuccessful(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(*testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		helpers.SeedAccountWithProfile(ctx, t, testConn, "buddy@santa.com", "anotherpassword", "Buddy", "TheElf")

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s", appPort))
		if err != nil {
			t.Fatalf("could not goto: %v", err)
		}

		err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
		if err != nil {
			t.Fatal("failed to click dropown button")
		}

		err = page.Locator("text=Sign In").Click()
		if err != nil {
			t.Fatal("failed to click link to sign in")
		}

		curURL := page.URL()
		if !strings.Contains(curURL, "/login") {
			t.Fatalf("expected to be on login page, got %s", curURL)
		}

		helpers.FillInField(t, "Email Address", "buddy@santa.com", page)
		helpers.FillInField(t, "Password", "anotherpassword", page)

		err = page.Locator("button:has-text('Sign In')").Click()
		if err != nil {
			t.Fatalf("could not click Sign In button: %v", err)
		}

		curURL = page.URL()
		if !strings.Contains(curURL, "/profile") {
			t.Fatalf("expected to be on profile page, got %s", curURL)
		}

		err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
		if err != nil {
			t.Fatal("failed to click dropown button")
		}

		locatorChecker := playwright.NewPlaywrightAssertions()

		dropdownMenu := page.Locator("#user-dropdown")
		dropdownAsserter := locatorChecker.Locator(dropdownMenu)

		err = dropdownAsserter.ToContainText("Sign Out")
		if err != nil {
			t.Fatalf("expected dropdown menu to contain 'Sign Out', got %v", err)
		}

		err = dropdownAsserter.ToContainText("Buddy TheElf")
		if err != nil {
			t.Fatalf("expected dropdown menu to contain 'Buddy TheElf', got %v", err)
		}

		err = dropdownAsserter.ToContainText("Settings")
		if err != nil {
			t.Fatalf("expected dropdown menu to contain 'Settings', got %v", err)
		}
	}
}

func testLoginFailsWhenUsernameOrPasswordIsIncorrect(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(*testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		helpers.SeedAccountWithProfile(ctx, t, testConn, "buddy@santa.com", "anotherpassword", "Buddy", "TheElf")

		testCases := map[string]struct {
			email    string
			password string
		}{
			"email only incorrect": {
				email:    "WRONG@gmail.com",
				password: "anotherpassword",
			},
			"password only incorrect": {
				email:    "buddy@santa.com",
				password: "WRONG",
			},
			"both incorrect": {
				email:    "WRONG@santa.com",
				password: "WRONG",
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				_, err := page.Goto(fmt.Sprintf("http://localhost:%s", appPort))
				if err != nil {
					t.Fatalf("could not goto: %v", err)
				}

				err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
				if err != nil {
					t.Fatal("failed to click dropown button")
				}

				err = page.Locator("text=Sign In").Click()
				if err != nil {
					t.Fatal("failed to click link to sign in")
				}

				curURL := page.URL()
				if !strings.Contains(curURL, "/login") {
					t.Fatalf("expected to be on login page, got %s", curURL)
				}

				helpers.FillInField(t, "Email Address", tc.email, page)
				helpers.FillInField(t, "Password", tc.password, page)

				err = page.Locator("button:has-text('Sign In')").Click()
				if err != nil {
					t.Fatalf("could not click Sign In button: %v", err)
				}

				curURL = page.URL()
				if !strings.Contains(curURL, "/login") {
					t.Fatalf("expected to be on login page, got %s", curURL)
				}

				locatorChecker := playwright.NewPlaywrightAssertions()

				flashMsg := page.GetByText("Email/Password combination is incorrect")
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
			})
		}
	}
}
