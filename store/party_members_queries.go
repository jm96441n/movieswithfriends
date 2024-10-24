package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const AddFriendToPartyQuery = `insert into party_members (id_member, id_party) values($1, $2)`

func (p *PGStore) CreatePartyMember(ctx context.Context, idMember, idParty int) error {
	_, err := p.db.Exec(ctx, AddFriendToPartyQuery, idMember, idParty)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgUniqueViolationCode {
				return ErrMemberPartyCombinationNotUnique
			}
		}
		return err
	}
	return nil
}
