package partymgmt

import (
	"context"
	"errors"
	"math/rand"

	"github.com/jm96441n/movieswithfriends/store"
)

type partyStore interface {
	CreateParty(context.Context, int, string, string) (int, error)
	GetPartyByShortID(context.Context, string) (store.Party, error)
	CreateProfileParty(context.Context, int, int) error
}

type PartyService struct {
	DB partyStore
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (s *PartyService) AddFriendToParty(ctx context.Context, idProfile int, shortID string) error {
	party, err := s.DB.GetPartyByShortID(ctx, shortID)
	if err != nil {
		return err
	}

	err = s.DB.CreateProfileParty(ctx, idProfile, party.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PartyService) CreateParty(ctx context.Context, idProfile int, name string) (int, error) {
	successFullyCreated := false
	var (
		id  int
		err error
	)
	for i := 0; i < 5; i++ {
		shortID := generateRandomString()
		id, err = s.DB.CreateParty(ctx, idProfile, name, shortID)
		if errors.Is(err, store.ErrDuplicatePartyShortID) {
			continue
		}
		if err != nil {
			return 0, err
		}

		successFullyCreated = true
		break
	}

	if !successFullyCreated {
		return 0, errors.New("failed to create party")
	}

	return id, nil
}

// generate a random 6 character string
func generateRandomString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
