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

		// Check the counts for movies watched
		helpers.LocatorHasText(t, page, pageAssertions, "#count-watched-movies", "8") // 2 movies from party1, 3 movies from party2, 3 movies from party3 -> 8

		// Check the counts for joined parties
		helpers.LocatorHasText(t, page, pageAssertions, "#count-joined-parties", "3") // 3 parties joined

		// Check the watch time
		helpers.LocatorHasText(t, page, pageAssertions, "#watch-time", "14h 52m") // (2 movies * 125 minutes each) + (3 movies * 120 minutes each) + (3 movies * 94 minutes each) = 250min + 360min + 282min = 892min = 14h 52m

		// Check the number of parties being shown
		parties := page.Locator(".party-card")
		helpers.Assert(t, parties != nil, "could not find items in .party-card")

		err = pageAssertions.Locator(parties).ToHaveCount(4) // 3 parties and 1 section for the "Create New Party" card
		helpers.Ok(t, err, "expected 3 parties and the 'Create Party' card, got %v", err)

		// check that we paginate watched movies to 5 in a list, we should see 5 currently and when we hit the next page button we should see 3
		recentlyWatchedMovies := page.Locator(".recently-watched-movie")
		helpers.Assert(t, recentlyWatchedMovies != nil, "could not find recently watched movies in .recently-watched-movie")

		err = pageAssertions.Locator(recentlyWatchedMovies).ToHaveCount(5)
		helpers.Ok(t, err, "expected 5 recently watched movies, got %v", err)
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
			MovieRuntime:     125,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       3,
			NumMovies:        9,
			NumWatchedMovies: 3,
			MovieRuntime:     120,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       1,
			NumMovies:        5,
			NumWatchedMovies: 3,
			MovieRuntime:     94,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       2,
			NumMovies:        5,
			NumWatchedMovies: 3,
			MovieRuntime:     120,
		},
	}

	for _, partyCfg := range partyCfgs {
		helpers.SeedPartyWithUsersAndMovies(ctx, t, testConn, partyCfg)
	}
}
