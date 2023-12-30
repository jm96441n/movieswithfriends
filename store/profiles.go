package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID        int     `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Parties   []Party `json:"parties"`
}

const GetProfileByIDQuery = `select id_profile, first_name, last_name from profiles where id_profile = $1`

func (pg *PGStore) GetProfileByID(ctx context.Context, id int) (Profile, error) {
	profile := Profile{}

	err := pg.db.QueryRow(ctx, GetProfileByIDQuery, id).Scan(&profile.ID, &profile.FirstName, &profile.LastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrNoRecord
		}

		return Profile{}, err
	}

	parties, err := pg.GetPartiesForProfile(ctx, id)
	if err != nil {
		return Profile{}, err
	}

	profile.Parties = parties

	return profile, nil
}
