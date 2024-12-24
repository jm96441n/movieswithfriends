package e2e_test

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jm96441n/movieswithfriends/e2e/helpers"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestSignup(t *testing.T) {
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

		page := helpers.OpenPage(t, browser, fmt.Sprintf("http://localhost:%s/signup", appPort))

		helpers.FillInField(t, "First Name", "Buddy", page)
		helpers.FillInField(t, "Last Name", "TheElf", page)
		helpers.FillInField(t, "Email", "buddy3@santa.com", page)
		helpers.FillInField(t, "Password", "1Password", page)

		err := page.GetByText("Create Account").Click()
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

		helpers.FillInField(t, "Email Address", "buddy3@santa.com", page)
		helpers.FillInField(t, "Password", "1Password", page)

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
		testConn := helpers.SetupDBConn(ctx, t, dbCtr)
		t.Cleanup(helpers.CleanupAndResetDB(ctx, t, dbCtr, testConn))

		helpers.SeedAccount(ctx, t, testConn, "buddy@santa.com", "anotherpassword")

		page := helpers.OpenPage(t, browser, fmt.Sprintf("http://localhost:%s/signup", appPort))

		helpers.FillInField(t, "First Name", "Buddy", page)
		helpers.FillInField(t, "Last Name", "TheElf", page)
		helpers.FillInField(t, "Email", "buddy@santa.com", page)
		helpers.FillInField(t, "Password", "1Password", page)

		err := page.GetByText("Create Account").Click()
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
		page := helpers.OpenPage(t, browser, fmt.Sprintf("http://localhost:%s/signup", appPort))

		expectedMsgs := []string{"First Name is required", "Last Name is required", "Email is required", "Password must contain:\nAt least 8 characters\nAt least one uppercase letter\nAt least one lowercase letter\nAt least one number"}
		fieldVals := [][]string{
			{"First Name", "Buddy"},
			{"Last Name", "TheElf"},
			{"Email", "buddy3@santa.com"},
			{"Passowrd", "1Password"},
		}

		errMsgs := page.Locator(".invalid-feedback:visible")
		if errMsgs == nil {
			t.Fatalf("could not get error messages")
		}

		err := playwright.NewPlaywrightAssertions().Locator(errMsgs).ToHaveCount(0)
		if err != nil {
			t.Fatal("Expected 0 warnings before submission but there were some")
		}

		for i := 0; i < len(expectedMsgs); i++ {
			err := page.GetByText("Create Account").Click()
			if err != nil {
				t.Fatalf("could not click create account button: %v", err)
			}

			errMsgs := page.Locator(".invalid-feedback:visible")
			if errMsgs == nil {
				t.Fatalf("could not get error messages")
			}

			err = playwright.NewPlaywrightAssertions().Locator(errMsgs).ToHaveCount(len(expectedMsgs))
			if err != nil {
				t.Fatal("Expected 0 warnings before submission but there were some")
			}

			texts, err := errMsgs.AllInnerTexts()
			if err != nil {
				t.Fatalf("could not get error messages text: %v", err)
			}

			if !slices.Equal(texts, expectedMsgs) {
				t.Fatalf("expected error messages to be %v, got %v", expectedMsgs, texts)
			}

			helpers.FillInField(t, fieldVals[i][0], fieldVals[i][1], page)
			expectedMsgs = expectedMsgs[1:]
		}
	}
}
