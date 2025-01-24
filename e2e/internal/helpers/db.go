package helpers

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type TestAccountInfo struct {
	AccountID int
	ProfileID int
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func SeedAccountWithProfile(ctx context.Context, t *testing.T, conn *pgxpool.Pool, accountInfo TestAccountInfo) TestAccountInfo {
	t.Helper()

	txn, err := conn.Begin(ctx)
	Ok(t, err, "Failed to open transaction to seed account")

	defer txn.Rollback(ctx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(accountInfo.Password), bcrypt.DefaultCost)
	Ok(t, err, "failed to hash password")

	err = txn.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) returning id_account", accountInfo.Email, hashedPassword).Scan(&accountInfo.AccountID)
	Ok(t, err, "failed to insert account")

	err = txn.QueryRow(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3) returning id_profile", accountInfo.FirstName, accountInfo.LastName, accountInfo.AccountID).Scan(&accountInfo.ProfileID)
	Ok(t, err, "failed to insert profile")

	Ok(t, txn.Commit(ctx), "failed to commit transaction")

	return accountInfo
}

type PartyConfig struct {
	NumMembers       int
	NumMovies        int
	NumWatchedMovies int
	CurrentAccount   TestAccountInfo
	MovieRuntime     int
	CurrentUserOwns  bool
}

var partyNumber = atomic.Int64{}

func SeedPartyWithUsersAndMovies(ctx context.Context, t *testing.T, conn *pgxpool.Pool, cfg PartyConfig) (string, int) {
	t.Helper()

	accountInfos := make([]TestAccountInfo, 0, cfg.NumMembers+1)
	for i := 0; i < cfg.NumMembers; i++ {
		accountInfo := TestAccountInfo{
			Email:     fmt.Sprintf("random%d%d@gmail.com", i, partyNumber.Load()),
			Password:  "password",
			FirstName: "Random",
			LastName:  "LastName",
		}
		accountInfo = SeedAccountWithProfile(ctx, t, conn, accountInfo)
		accountInfos = append(accountInfos, accountInfo)
	}

	// CurrentAccount might be an empty object, if that's the case then don't append it, this means the current user is not part of this party
	if cfg.CurrentAccount.AccountID != 0 {
		accountInfos = append(accountInfos, cfg.CurrentAccount)
	}

	movies := make([]int, 0, cfg.NumMovies)
	for i := 0; i < cfg.NumMovies; i++ {
		movies = append(movies, SeedMovie(ctx, t, conn, fmt.Sprintf("Movie %d", i), cfg.MovieRuntime))
	}
	partyName := fmt.Sprintf("party name %d", partyNumber.Add(1))
	partyID := SeedParty(ctx, t, conn, partyName)

	owned := false
	for _, accountInfo := range accountInfos {
		// if the accountInfo we're on is the current account and the currnet account owns the party then set the current account as the owner
		// OR if the party is not owned yet and the current user does not own the party then set the owner to be the first account we come across
		if (cfg.CurrentAccount.AccountID == accountInfo.AccountID && cfg.CurrentUserOwns) || (!owned && !cfg.CurrentUserOwns) {
			addMemberToPartyAsOwner(ctx, t, conn, partyID, accountInfo.AccountID)
			owned = true
			continue
		}

		addMemberToParty(ctx, t, conn, partyID, accountInfo.AccountID)
	}

	for idx, movieID := range movies {
		accountAddedBy := accountInfos[idx%len(accountInfos)]
		if cfg.NumWatchedMovies > 0 {
			addWatchedMovieToParty(ctx, t, conn, partyID, movieID, accountAddedBy.AccountID)
			cfg.NumWatchedMovies--
			continue
		}
		addMovieToParty(ctx, t, conn, partyID, movieID, accountAddedBy.AccountID)
	}
	return partyName, partyID
}

const insertMovieQuery = `INSERT INTO movies (title, poster_url, tmdb_id, overview, tagline, runtime) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id_movie`

func SeedMovie(ctx context.Context, t *testing.T, conn *pgxpool.Pool, title string, runtime int) int {
	t.Helper()

	var movieID int
	err := conn.QueryRow(ctx, insertMovieQuery, title, "poster.com", 12345, "it's a movie", "still a movie", runtime).Scan(&movieID)
	Ok(t, err, "failed to insert movie")

	return movieID
}

func SeedParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, name string) int {
	t.Helper()

	var partyID int
	err := conn.QueryRow(ctx, `INSERT INTO parties (name) VALUES ($1) RETURNING id_party`, name).Scan(&partyID)
	Ok(t, err, "failed to insert party")

	return partyID
}

func addMemberToParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID, accountID int) {
	t.Helper()

	_, err := conn.Exec(ctx, `INSERT INTO party_members (id_party, id_member) VALUES ($1, $2)`, partyID, accountID)
	Ok(t, err, "failed to add member to party")
}

func addMemberToPartyAsOwner(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID, accountID int) {
	t.Helper()

	_, err := conn.Exec(ctx, `INSERT INTO party_members (id_party, id_member, owner) VALUES ($1, $2, true)`, partyID, accountID)
	Ok(t, err, "failed to add member to party")
}

func addMovieToParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID, movieID, accountID int) {
	t.Helper()

	_, err := conn.Exec(ctx, `INSERT INTO party_movies (id_party, id_movie, id_added_by) VALUES ($1, $2, $3)`, partyID, movieID, accountID)
	Ok(t, err, "failed to add movie to party")
}

func addWatchedMovieToParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID, movieID, accountID int) {
	t.Helper()

	_, err := conn.Exec(ctx, `INSERT INTO party_movies (id_party, id_movie, id_added_by, watch_status, watch_date) VALUES ($1, $2, $3, 'watched', $4)`, partyID, movieID, accountID, time.Now())
	Ok(t, err, "failed to add watched movie to party")
}
