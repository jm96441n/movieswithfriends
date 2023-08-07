package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID        int
	Name      string
	AccountID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const GetProfileQuery = `SELECT * FROM profiles
WHERE profiles.id = $1`

func (pg *PGStore) GetProfile(ctx context.Context, id int) (Profile, error) {
	rows, err := pg.db.Query(ctx, GetProfileQuery, id)
	if err != nil {
		return Profile{}, err
	}
	profile, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Profile])
	if err != nil {
		return Profile{}, err
	}

	return profile, nil
}
