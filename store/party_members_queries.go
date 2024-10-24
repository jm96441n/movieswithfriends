package store

import "context"

const AddFriendToPartyQuery = `insert into party_members (id_member, id_party) values($1, $2)`

func (p *PGStore) CreatePartyMember(ctx context.Context, idMember, idParty int) error {
	_, err := p.db.Exec(ctx, AddFriendToPartyQuery, idMember, idParty)
	return err
}
