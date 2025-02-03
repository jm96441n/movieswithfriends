package store

import "github.com/jackc/pgx/v5/pgxpool"

type InvitationsRepository struct {
	db *pgxpool.Pool
}

func NewInvitationsRepository(db *pgxpool.Pool) InvitationsRepository {
	return InvitationsRepository{db: db}
}
