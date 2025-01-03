package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func testCreatePartyIsSuccessful(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(*testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		accountInfo := helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "buddy@santa.com", Password: "anotherpassword", FirstName: "Buddy", LastName: "TheElf"})
		helpers.LoginAs(t, page, accountInfo)

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/parties/new", appPort))
		helpers.Ok(t, err, "could not goto create party page")

		helpers.FillInField(t, helpers.FormField{Label: "Party Name", Value: "My Party"}, page)

		helpers.Ok(t, page.Locator("button:has-text('Create Party')").Click(), "failed to click create party button")

		time.Sleep(2 * time.Second)
	}
}
