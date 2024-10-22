package store

import "context"

const AddFriendToPartyQuery = `insert into profile_parties (id_profile, id_party) values($1, $2)`

func (p *PGStore) CreateProfileParty(ctx context.Context, idProfile, idParty int) error {
	_, err := p.db.Exec(ctx, AddFriendToPartyQuery, idProfile, idParty)
	return err
}
