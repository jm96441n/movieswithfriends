package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AccountID int
}

const GetProfileByIDQuery = `
  select profiles.id_profile, profiles.first_name, profiles.last_name from profiles 
  where profiles.id_profile = $1`

func (pg *PGStore) GetProfileByID(ctx context.Context, profileID int) (Profile, error) {
	profile := Profile{}

	err := pg.db.QueryRow(ctx, GetProfileByIDQuery, profileID).Scan(&profile.ID, &profile.FirstName, &profile.LastName)
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
