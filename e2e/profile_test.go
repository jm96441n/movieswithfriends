package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func TestProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(*testing.T){
		"testCanViewAndEditProfile": testCanViewAndEditProfile(ctx, connPool, page, port),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testCanViewAndEditProfile(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		currentAccount := helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "buddy@santa.com", Password: "anotherpassword", FirstName: "Buddy", LastName: "TheElf"})
		setupProfileViewData(ctx, t, testConn, currentAccount)
		helpers.LoginAs(t, page, currentAccount)

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/profile", appPort))
		helpers.Ok(t, err, "could not goto profile page")

		pageAssertions := playwright.NewPlaywrightAssertions()

		// Check the number of parties being shown
		parties := page.Locator(".party-card")
		helpers.Assert(t, parties != nil, "could not get error messages")

		err = pageAssertions.Locator(parties).ToHaveCount(4) // 3 parties and 1 section for the "Create New Party" card
		helpers.Ok(t, err, "expected 3 parties and the 'Create Party' card, got %v", err)

		// Check the counts for movies watched
		movieWatchCount := page.Locator("#count-movies-watched")
		helpers.Assert(t, parties != nil, "could not get error messages")

		err = pageAssertions.Locator(movieWatchCount).ToHaveText("8") // 2 movies from party1, 3 movies from party2, 3 movies from party3 -> 8
		helpers.Ok(t, err, "expected 8 movies watched, got %v", err)
	}
}

func setupProfileViewData(ctx context.Context, t *testing.T, testConn *pgxpool.Pool, currentAccount helpers.TestAccountInfo) {
	// user not in any party
	helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "anotherUser@gmail.com", Password: "anotherpassword", FirstName: "Another", LastName: "User"})
	partyCfgs := []helpers.PartyConfig{
		{
			NumMembers:       2,
			NumMovies:        8,
			NumWatchedMovies: 2,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       3,
			NumMovies:        9,
			NumWatchedMovies: 3,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       1,
			NumMovies:        5,
			NumWatchedMovies: 3,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       2,
			NumMovies:        5,
			NumWatchedMovies: 3,
		},
	}

	for _, partyCfg := range partyCfgs {
		helpers.SeedPartyWithUsersAndMovies(ctx, t, testConn, partyCfg)
	}
}
