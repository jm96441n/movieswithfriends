package partymgmt

import (
	"context"
	"time"

	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type InvitationsService struct {
	db store.InvitationsRepository
}

func NewInvitationsService(db store.InvitationsRepository) InvitationsService {
	return InvitationsService{db: db}
}

type Invite struct {
	Email      string
	InviteDate time.Time
}

func (i InvitationsService) GetInvitationsForParty(ctx context.Context, idParty int) ([]Invite, error) {
	return nil, nil
}
