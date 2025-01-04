package e2e_test

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func TestSignup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(*testing.T){
		"testSignupIsSuccessful":             testSignupIsSuccessful(ctx, connPool, page, port),
		"testSignupFailsIfEmailIsInUse":      testSignupFailsIfEmailIsInUse(ctx, connPool, page, port),
		"testSignupFailsWithFormValidations": testSignupFailsWithFormValidations(ctx, connPool, page, port),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testSignupIsSuccessful(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort))
		helpers.Ok(t, err, "could not goto signup page")

		helpers.FillInField(t, helpers.FormField{Label: "First Name", Value: "Buddy"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Last Name", Value: "TheElf"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Email", Value: "buddy3@santa.com"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Password", Value: "1Password"}, page)

		err = page.Locator("button:has-text('Create Account')").Click()
		helpers.Ok(t, err, "could not click create account button")

		flashMsg := page.GetByText("Successfully signed up! Please log in.")
		helpers.Assert(t, flashMsg != nil, "expected success message, got nil")

		curURL := page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/login"), "expected to be on login page, got %s", curURL)

		helpers.FillInField(t, helpers.FormField{Label: "Email Address", Value: "buddy3@santa.com"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Password", Value: "1Password"}, page)

		err = page.Locator("button:has-text('Sign In')").Click()
		helpers.Ok(t, err, "could not click sign in button")

		curURL = page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/profile"), "expected to be on profile page, got %s", curURL)
	}
}

func testSignupFailsIfEmailIsInUse(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)

		helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "buddy@santa.com", Password: "anotherpassword", FirstName: "Buddy", LastName: "TheElf"})

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort))
		helpers.Ok(t, err, "could not goto signup page")

		helpers.FillInField(t, helpers.FormField{Label: "First Name", Value: "Buddy"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Last Name", Value: "TheElf"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Email", Value: "buddy@santa.com"}, page)
		helpers.FillInField(t, helpers.FormField{Label: "Password", Value: "1Password"}, page)

		err = page.Locator("button:has-text('Create Account')").Click()
		helpers.Ok(t, err, "could not click create account button")

		locatorChecker := playwright.NewPlaywrightAssertions()

		flashMsg := page.GetByText("An account exists with this email. Try logging in or resetting your password.")
		flashChecker := locatorChecker.Locator(flashMsg)
		helpers.Ok(t, flashChecker.Not().ToBeEmpty(), "could not get flash message")

		regex := regexp.MustCompile(`.*alert-danger.*`)

		if err := flashChecker.ToHaveClass(regex); err != nil {
			s, err := flashMsg.GetAttribute("class")
			helpers.Ok(t, err, "could not get class attribute")
			t.Fatalf("expected flash message to be have class \"alert-danger\", it was %s", s)
		}

		curURL := page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/signup"), "expected to be on signup page, got %s", curURL)
	}
}

func testSignupFailsWithFormValidations(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/signup", appPort))
		helpers.Ok(t, err, "could not goto signup page")

		expectedMsgs := []string{"First Name is required", "Last Name is required", "Email is required", "Password must contain:\nAt least 8 characters\nAt least one uppercase letter\nAt least one lowercase letter\nAt least one number"}
		fieldVals := [][]string{
			{"First Name", "Buddy"},
			{"Last Name", "TheElf"},
			{"Email", "buddy3@santa.com"},
			{"Password", "1Password"},
		}

		errMsgs := page.Locator(".invalid-feedback:visible")
		helpers.Assert(t, errMsgs != nil, "could not get error messages")

		err = playwright.NewPlaywrightAssertions().Locator(errMsgs).ToHaveCount(0)
		helpers.Ok(t, err, "Expected 0 warnings before submission but there were some")

		for i := 0; i < 4; i++ {
			err := page.Locator("button:has-text('Create Account')").Click()
			helpers.Ok(t, err, "could not click create account button")

			errMsgs := page.Locator(".invalid-feedback:visible")
			helpers.Assert(t, errMsgs != nil, "could not get error messages")

			err = playwright.NewPlaywrightAssertions().Locator(errMsgs).ToHaveCount(len(expectedMsgs))
			helpers.Ok(t, err, "Expected %d warnings before submission but there were not", len(expectedMsgs))

			texts, err := errMsgs.AllInnerTexts()
			helpers.Ok(t, err, "could not get error messages text")

			helpers.Assert(t, slices.Equal(texts, expectedMsgs), "expected error messages to be %v, got %v", expectedMsgs, texts)

			helpers.FillInField(t, helpers.FormField{Label: fieldVals[i][0], Value: fieldVals[i][1]}, page)
			expectedMsgs = expectedMsgs[1:]
		}
	}
}
