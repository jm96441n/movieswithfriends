package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	db *pgxpool.Pool
}

const pgUniqueViolationCode = "23505"

var (
	ErrNoRecord              = errors.New("store: no matching record found")
	ErrDuplicateEmailAddress = errors.New("email address already exists")
)

func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

type GetProfileResult struct {
	ID              int
	FirstName       string
	LastName        string
	CreatedAt       time.Time
	AccountID       int
	AccountEmail    string
	AccountPassword []byte
}

const getProfileByIDQuery = `
  select 
    profiles.id_profile,
    profiles.first_name,
    profiles.last_name,
    profiles.created_at,
    accounts.id_account,
    accounts.email,
    accounts.password
  from profiles
  join accounts on profiles.id_account = accounts.id_account
  where profiles.id_profile = $1`

func (p *ProfileRepository) GetProfileByID(ctx context.Context, profileID int) (GetProfileResult, error) {
	res := GetProfileResult{}

	err := p.db.QueryRow(ctx, getProfileByIDQuery, profileID).
		Scan(&res.ID, &res.FirstName, &res.LastName, &res.CreatedAt, &res.AccountID, &res.AccountEmail, &res.AccountPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetProfileResult{}, ErrNoRecord
		}

		return GetProfileResult{}, err
	}

	return res, nil
}

const getProfileByEmailQuery = `
  select 
    profiles.id_profile,
    profiles.first_name,
    profiles.last_name,
    profiles.created_at,
    accounts.id_account,
    accounts.email,
  accounts.password
  from profiles
  join accounts on profiles.id_account = accounts.id_account
  where accounts.email = $1`

func (p *ProfileRepository) GetProfileByEmail(ctx context.Context, email string) (GetProfileResult, error) {
	res := GetProfileResult{}
	err := p.db.QueryRow(ctx, getProfileByEmailQuery, email).
		Scan(&res.ID, &res.FirstName, &res.LastName, &res.CreatedAt, &res.AccountID, &res.AccountEmail, &res.AccountPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return GetProfileResult{}, ErrNoRecord
		}
		return GetProfileResult{}, err
	}
	return res, nil
}

type GetProfileStatsResult struct {
	NumParties    int
	WatchTime     int
	MoviesWatched int
}

const getNumPartiesForProfileQuery = `select count(*) from party_members where id_member = $1;`

const getTotalWatchTimeQuery = `
  select coalesce(sum(movies.runtime), 0), count(movies.*) from movies
  join party_movies on party_movies.id_movie = movies.id_movie
  join party_members on party_members.id_party = party_movies.id_party
  where party_members.id_member = $1 AND party_movies.watch_status = 'watched';
`

func (p *ProfileRepository) GetProfileStats(ctx context.Context, logger *slog.Logger, profileID int) (GetProfileStatsResult, error) {
	res := GetProfileStatsResult{}
	err := p.db.QueryRow(ctx, getNumPartiesForProfileQuery, profileID).Scan(&res.NumParties)
	if err != nil {
		logger.Debug("failed to get num parties for profile", slog.Any("error", err))
		return GetProfileStatsResult{}, err
	}
	err = p.db.QueryRow(ctx, getTotalWatchTimeQuery, profileID).Scan(&res.WatchTime, &res.MoviesWatched)
	if err != nil {
		logger.Debug("failed to get total watch time", slog.Any("error", err))
		return GetProfileStatsResult{}, err
	}
	return res, nil
}

type CreateProfileResult struct {
	AccountID int
	ProfileID int
}

func (p *ProfileRepository) CreateProfile(ctx context.Context, email, firstName, lastName string, password []byte) (CreateProfileResult, error) {
	txn, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateProfileResult{}, err
	}

	defer txn.Rollback(ctx)

	res := CreateProfileResult{}

	res.AccountID, err = p.createAccountWithTxn(ctx, txn, email, password)
	if err != nil {
		return CreateProfileResult{}, err
	}

	res.ProfileID, err = p.createProfileWithTxn(ctx, txn, firstName, lastName, res.AccountID)
	if err != nil {
		return CreateProfileResult{}, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return CreateProfileResult{}, err
	}

	return res, nil
}

const insertProfileQuery = `insert into profiles (first_name, last_name, id_account) values ($1, $2, $3) returning id_profile`

func (p *ProfileRepository) createProfileWithTxn(ctx context.Context, txn pgx.Tx, firstName, lastName string, accountID int) (int, error) {
	var profileID int

	err := txn.QueryRow(ctx, insertProfileQuery, firstName, lastName, accountID).Scan(&profileID)
	if err != nil {
		return 0, err
	}
	return profileID, nil
}

const insertAccountQuery = `insert into accounts (email, password) values ($1, $2) returning id_account`

func (p *ProfileRepository) createAccountWithTxn(ctx context.Context, txn pgx.Tx, email string, password []byte) (int, error) {
	var accountID int

	err := txn.QueryRow(ctx, insertAccountQuery, email, password).Scan(&accountID)
	if err != nil {
		return 0, handleUniqueConstraintForEmail(err)
	}

	return accountID, nil
}

const accountExistsQuery = `SELECT EXISTS(select true from accounts where id_account = $1)`

func (p *ProfileRepository) AccountExists(ctx context.Context, id int) (bool, error) {
	var exists bool

	err := p.db.QueryRow(ctx, accountExistsQuery, id).Scan(&exists)

	return exists, err
}

type AccountUpdateAttrs struct {
	ID       int
	Email    string
	Password []byte
}

type ProfileUpdateAttrs struct {
	ID        int
	FirstName string
	LastName  string
}

func (p *ProfileRepository) UpdateProfile(ctx context.Context, accountAttrs AccountUpdateAttrs, profileAttrs ProfileUpdateAttrs) error {
	txn, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer txn.Rollback(ctx)

	err = updateAccount(ctx, txn, accountAttrs)
	if err != nil {
		return err
	}

	err = updateProfile(ctx, txn, profileAttrs)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func updateAccount(ctx context.Context, txn pgx.Tx, attrs AccountUpdateAttrs) error {
	if len(attrs.Password) == 0 {
		_, err := txn.Exec(ctx, `update accounts set email = $1 where id_account = $2`, attrs.Email, attrs.ID)
		if err != nil {
			return handleUniqueConstraintForEmail(err)
		}
		return nil
	}

	_, err := txn.Exec(ctx, `update accounts set email = $1, password = $2 where id_account = $3`, attrs.Email, attrs.Password, attrs.ID)
	if err != nil {
		return handleUniqueConstraintForEmail(err)
	}

	return nil
}

func handleUniqueConstraintForEmail(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgUniqueViolationCode {
			if pgErr.ConstraintName == "unique_email_addresses" {
				return ErrDuplicateEmailAddress
			}
		}
	}

	return err
}

func updateProfile(ctx context.Context, txn pgx.Tx, attrs ProfileUpdateAttrs) error {
	_, err := txn.Exec(ctx, `update profiles set first_name = $1, last_name = $2 where id_profile = $3`, attrs.FirstName, attrs.LastName, attrs.ID)
	if err != nil {
		return err
	}

	return nil
}
