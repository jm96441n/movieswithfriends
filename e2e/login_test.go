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
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(t *testing.T){
		"testLoginIsSuccessful":                           testLoginIsSuccessful(ctx, connPool, page, port),
		"testLoginFailsWhenUsernameOrPasswordIsIncorrect": testLoginFailsWhenUsernameOrPasswordIsIncorrect(ctx, connPool, page, port),
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
		helpers.Ok(t, err, "could not goto index page")

		err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
		helpers.Ok(t, err, "failed to click dropown button")

		err = page.Locator("text=Sign In").Click()
		helpers.Ok(t, err, "failed to click link to sign in")

		curURL := page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/login"), "expected to be on login page, got %s", curURL)

		helpers.FillInField(t, "Email Address", "buddy@santa.com", page)
		helpers.FillInField(t, "Password", "anotherpassword", page)

		err = page.Locator("button:has-text('Sign In')").Click()
		helpers.Ok(t, err, "could not click Sign In button")

		curURL = page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/profile"), "expected to be on profile page, got %s", curURL)

		err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
		helpers.Ok(t, err, "failed to click dropown button")

		locatorChecker := playwright.NewPlaywrightAssertions()

		dropdownMenu := page.Locator("#user-dropdown")
		dropdownAsserter := locatorChecker.Locator(dropdownMenu)

		err = dropdownAsserter.ToContainText("Sign Out")
		helpers.Ok(t, err, "expected dropdown menu to contain 'Sign Out', got %v", err)

		err = dropdownAsserter.ToContainText("Buddy TheElf")
		helpers.Ok(t, err, "expected dropdown menu to contain 'Buddy TheElf', got %v", err)

		err = dropdownAsserter.ToContainText("Settings")
		helpers.Ok(t, err, "expected dropdown menu to contain 'Settings', got %v", err)
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
				helpers.Ok(t, err, "could not goto index page")

				err = page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
				helpers.Ok(t, err, "failed to click dropown button")

				err = page.Locator("text=Sign In").Click()
				helpers.Ok(t, err, "failed to click link to sign in")

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
				helpers.Assert(t, strings.Contains(curURL, "/login"), "expected to be on login page, got %s", curURL)

				locatorChecker := playwright.NewPlaywrightAssertions()

				flashMsg := page.GetByText("Email/Password combination is incorrect")
				flashChecker := locatorChecker.Locator(flashMsg)
				helpers.Ok(t, flashChecker.Not().ToBeEmpty(), "expected error message in flash, got nothing")

				regex := regexp.MustCompile(`.*alert-danger.*`)

				err = flashChecker.ToHaveClass(regex)
				if err != nil {
					s, err := flashMsg.GetAttribute("class")
					helpers.Ok(t, err, "failed to get class attribute")

					t.Fatalf("expected flash message to be have class \"alert-danger\", it was %s", s)
				}
			})
		}
	}
}
