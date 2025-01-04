package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		time.Sleep(3 * time.Second)
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
