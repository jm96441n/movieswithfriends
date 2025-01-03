package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	AccountID int
}

const GetProfileByIDQuery = `
  select 
    profiles.id_profile,
    profiles.first_name,
    profiles.last_name,
    profiles.id_account,
    profiles.created_at
  from profiles
  where profiles.id_profile = $1`

func (pg *PGStore) GetProfileByID(ctx context.Context, profileID int) (Profile, error) {
	profile := Profile{}

	err := pg.db.QueryRow(ctx, GetProfileByIDQuery, profileID).Scan(&profile.ID, &profile.FirstName, &profile.LastName, &profile.AccountID, &profile.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrNoRecord
		}

		return Profile{}, err
	}

	return profile, nil
}

const insertProfileQuery = `insert into profiles (first_name, last_name, id_account) values ($1, $2, $3) returning id_profile`

func (pg *PGStore) CreateProfileWithTxn(ctx context.Context, txn pgx.Tx, firstName, lastName string, accountID int) (Profile, error) {
	profile := Profile{FirstName: firstName, LastName: lastName, AccountID: accountID}

	err := txn.QueryRow(ctx, insertProfileQuery, firstName, lastName, accountID).Scan(&profile.ID)
	if err != nil {
		return Profile{}, err
	}
	return profile, nil
}

const getNumPartiesForProfileQuery = `select count(*) from party_members where id_member = $1;`

const getTotalWatchTimeQuery = `
  select coalesce(sum(movies.runtime), 0), count(movies.*) from movies
  join party_movies on party_movies.id_movie = movies.id_movie
  join party_members on party_members.id_party = party_movies.id_party
  where party_members.id_member = $1 AND party_movies.watch_status = 'watched';
`

func (pg *PGStore) GetProfileStats(ctx context.Context, profileID int) (int, int, int, error) {
	var numParties, watchTime, moviesWatched int
	err := pg.db.QueryRow(ctx, getNumPartiesForProfileQuery, profileID).Scan(&numParties)
	if err != nil {
		pg.logger.Error("failed to get num parties for profile", slog.Any("error", err))
		return 0, 0, 0, err
	}
	err = pg.db.QueryRow(ctx, getTotalWatchTimeQuery, profileID).Scan(&watchTime, &moviesWatched)
	if err != nil {
		pg.logger.Error("failed to get total watch time", slog.Any("error", err))
		return 0, 0, 0, err
	}
	return numParties, watchTime, moviesWatched, nil
}

func (pg *PGStore) GetAccountEmailForProfile(ctx context.Context, profileID int) (string, error) {
	var email string
	err := pg.db.QueryRow(ctx, `select accounts.email from accounts join profiles on accounts.id_account = profiles.id_account where profiles.id_profile = $1`, profileID).Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}
