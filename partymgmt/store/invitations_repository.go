package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: this should be in the party repository
type InvitationsRepository struct {
	db *pgxpool.Pool
}

func NewInvitationsRepository(db *pgxpool.Pool) InvitationsRepository {
	return InvitationsRepository{db: db}
}

type invitationAssignFunc func(string, time.Time)

const getInvitationsForPartyQuery = `
  SELECT email, created_at
  FROM invitations
  WHERE id_party = $1`

func (i InvitationsRepository) GetInvitationsForParty(ctx context.Context, idParty int, assignFn invitationAssignFunc) error {
	rows, err := i.db.Query(ctx, getInvitationsForPartyQuery, idParty)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			email     string
			invitedAt time.Time
		)
		err = rows.Scan(&email, &invitedAt)
		if err != nil {
			return err
		}

		assignFn(email, invitedAt)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

const createInviteQuery = `
INSERT INTO invitations (id_party, email) VALUES ($1, $2)
`

func (i InvitationsRepository) CreateInviteWatcherDoesNotExist(ctx context.Context, idParty int, email string) error {
	_, err := i.db.Exec(ctx, createInviteQuery, idParty, email)
	return err
}

const createInviteForWatcherQuery = `  
INSERT INTO invitations (id_party, id_profile, email) VALUES ($1, $2, $3)
`

func (i InvitationsRepository) CreateInviteForWatcher(ctx context.Context, idParty, idWatcher int, email string) error {
	_, err := i.db.Exec(ctx, createInviteForWatcherQuery, idParty, idWatcher, email)
	return err
}
