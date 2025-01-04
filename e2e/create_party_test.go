package e2e_test

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

func TestCreateParty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(t *testing.T){
		"testCreatePartyIsSuccessful": testCreatePartyIsSuccessful(ctx, connPool, page, port),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

var partyPathRegex = regexp.MustCompile(`localhost:\d+/parties/\d+`)

func testCreatePartyIsSuccessful(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(*testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		accountInfo := helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "buddy@santa.com", Password: "anotherpassword", FirstName: "Buddy", LastName: "TheElf"})
		helpers.LoginAs(t, page, accountInfo)

		cases := map[string]struct {
			startPage        string
			createButtonText string
			partyName        string
		}{
			"profilePage": {
				startPage:        "profile",
				createButtonText: "Create Party",
				partyName:        "From Profile Page Party",
			},
			"partiesPage-CreatePartyButton": {
				startPage:        "parties",
				createButtonText: "Create Party",
				partyName:        "From Parties Page Create Party Button",
			},
			"partiesPage-CreateNewPartyButton": {
				startPage:        "parties",
				createButtonText: "Create New Party",
				partyName:        "From Parties Page Create New Party Button",
			},
		}

		for name, testCase := range cases {
			t.Run(name, func(t *testing.T) {
				_, err := page.Goto(fmt.Sprintf("http://localhost:%s/%s", appPort, testCase.startPage))
				helpers.Ok(t, err, "could not goto %s page", testCase.startPage)

				helpers.Assert(t, strings.Contains(page.URL(), fmt.Sprintf("/%s", testCase.startPage)), "expected to be on profile page, got %s", page.URL())

				helpers.Ok(t, page.GetByRole("link", playwright.PageGetByRoleOptions{Name: testCase.createButtonText}).Click(), "failed to click create party button on %s page", testCase.startPage)

				helpers.Assert(t, strings.Contains(page.URL(), "/parties/new"), "expected to be on /parties/new page, got %s", page.URL())

				helpers.FillInField(t, helpers.FormField{Label: "Party Name", Value: testCase.partyName}, page)

				helpers.Ok(t, page.Locator("button:has-text('Create Party')").Click(), "failed to click create party button on party create page")

				helpers.Assert(t, partyPathRegex.MatchString(page.URL()), "expected to be on party show page, got %s", page.URL())

				asserter := playwright.NewPlaywrightAssertions()
				helpers.Ok(t, asserter.Locator(page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: testCase.partyName})).ToBeVisible(), "expected to see party name on party show page, could not find it")

				flashMsg := page.GetByText("Party successfully created!")
				helpers.Assert(t, flashMsg != nil, "expected success message, got nil")
			})
		}
	}
}
