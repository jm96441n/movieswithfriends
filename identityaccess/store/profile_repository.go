package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

type GetProfileByIDResult struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreatedAt    time.Time `json:"created_at"`
	AccountEmail string
	AccountID    int
}

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
    profiles.created_at,
    accounts.email
    accounts.id_account
  from profiles
  join accounts on profiles.id_account = accounts.id_account
  where profiles.id_profile = $1`

var ErrNoRecord = errors.New("store: no matching record found")

func (p *ProfileRepository) GetProfileByID(ctx context.Context, profileID int) (GetProfileByIDResult, error) {
	profile := GetProfileByIDResult{}

	err := p.db.QueryRow(ctx, GetProfileByIDQuery, profileID).Scan(&profile.ID, &profile.FirstName, &profile.LastName, &profile.CreatedAt, &profile.AccountEmail, &profile.AccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetProfileByIDResult{}, ErrNoRecord
		}

		return GetProfileByIDResult{}, err
	}

	return profile, nil
}

const insertProfileQuery = `insert into profiles (first_name, last_name, id_account) values ($1, $2, $3) returning id_profile`

func (p *ProfileRepository) CreateProfileWithTxn(ctx context.Context, txn pgx.Tx, firstName, lastName string, accountID int) (Profile, error) {
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

func (p *ProfileRepository) GetProfileStats(ctx context.Context, logger *slog.Logger, profileID int) (int, int, int, error) {
	var numParties, watchTime, moviesWatched int
	err := p.db.QueryRow(ctx, getNumPartiesForProfileQuery, profileID).Scan(&numParties)
	if err != nil {
		logger.Debug("failed to get num parties for profile", slog.Any("error", err))
		return 0, 0, 0, err
	}
	err = p.db.QueryRow(ctx, getTotalWatchTimeQuery, profileID).Scan(&watchTime, &moviesWatched)
	if err != nil {
		logger.Debug("failed to get total watch time", slog.Any("error", err))
		return 0, 0, 0, err
	}
	return numParties, watchTime, moviesWatched, nil
}
