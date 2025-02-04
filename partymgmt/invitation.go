package partymgmt

import (
	"context"
	"errors"
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
	var invites []Invite
	err := i.db.GetInvitationsForParty(ctx, idParty, func(email string, inviteDate time.Time) {
		invites = append(invites, Invite{
			Email:      email,
			InviteDate: inviteDate,
		})
	})
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (i InvitationsService) CreateInvite(ctx context.Context, watcherService WatcherService, idParty int, email string) error {
	watcher, err := watcherService.GetWatcherByEmail(ctx, email)

	if err != nil && !errors.Is(err, ErrWatcherNotFound) {
		return err
	}

	// watcher does not exist yet so create invite without the reference
	if errors.Is(err, ErrWatcherNotFound) {
		err = i.db.CreateInviteWatcherDoesNotExist(ctx, idParty, email)
		if err != nil {
			return err
		}
		return nil
	}

	// watcher exists so create invite with the reference
	err = i.db.CreateInviteForWatcher(ctx, idParty, watcher.ID, email)
	if err != nil {
		return err
	}
	return nil
}
