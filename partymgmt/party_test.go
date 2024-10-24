package partymgmt_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jm96441n/movieswithfriends/partymgmt"
	"github.com/jm96441n/movieswithfriends/store"
)

func TestAddFriendToParty_HappyPath(t *testing.T) {
	t.Parallel()
	svc := &partymgmt.PartyService{
		DB: &mockPartyStore{},
	}

	idMember := 1
	err := svc.AddFriendToParty(context.Background(), idMember, "shortID")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAddFriendToParty_SadPath(t *testing.T) {
	errGetPartyByShortID := errors.New("error getting party by short id")
	errCreatePartyMember := errors.New("error creating party member")

	testCases := map[string]struct {
		errGetPartyByShortID error
		errCreatePartyMember error
		expectedError        error
	}{
		"db failed to get party by short id": {
			errGetPartyByShortID: errGetPartyByShortID,
			expectedError:        errGetPartyByShortID,
		},
		"db failed to create party member": {
			errCreatePartyMember: errCreatePartyMember,
			expectedError:        errCreatePartyMember,
		},
		"db failed to create party member because member exists in party": {
			errCreatePartyMember: store.ErrMemberPartyCombinationNotUnique,
			expectedError:        partymgmt.ErrMemberExistsInParty,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc := &partymgmt.PartyService{
				DB: &mockPartyStore{
					errGetPartyByShortID: tc.errGetPartyByShortID,
					errCreatePartyMember: tc.errCreatePartyMember,
				},
			}

			idMember := 1
			err := svc.AddFriendToParty(context.Background(), idMember, "shortID")
			if err == nil {
				t.Errorf("expected error, got nil")
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestCreateParty_HappyPath(t *testing.T) {
	testCases := map[string]struct {
		numTimesToErrOnCreateParty int
	}{
		"party created on first try": {},
		"party created on fourth try to generate unique short id": {
			numTimesToErrOnCreateParty: 4,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc := &partymgmt.PartyService{
				DB: &mockPartyStore{
					numTimesToErrOnCreateParty: tc.numTimesToErrOnCreateParty,
				},
			}
			idMember := 1
			name := "name"
			id, err := svc.CreateParty(context.Background(), idMember, name)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			if id == 0 {
				t.Errorf("expected id > 0, got %v", id)
			}
		})
	}
}

func TestCreateParty_SadPath(t *testing.T) {
	testCases := map[string]struct {
		numTimesToErrOnCreateParty int
		errOnCreateParty           bool
	}{
		"party takes more than 5 times to create short id": {
			numTimesToErrOnCreateParty: 6,
		},
		"db fails to create party": {
			errOnCreateParty: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc := &partymgmt.PartyService{
				DB: &mockPartyStore{
					numTimesToErrOnCreateParty: tc.numTimesToErrOnCreateParty,
					errOnCreateParty:           true,
				},
			}
			idMember := 1
			name := "name"
			_, err := svc.CreateParty(context.Background(), idMember, name)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

type mockPartyStore struct {
	errGetPartyByShortID error
	errCreatePartyMember error

	numTimesToErrOnCreateParty int
	errOnCreateParty           bool
}

func (m *mockPartyStore) CreateParty(ctx context.Context, idMember int, name string, shortID string) (int, error) {
	if m.numTimesToErrOnCreateParty > 0 {
		m.numTimesToErrOnCreateParty--
		return 0, store.ErrDuplicatePartyShortID
	}

	if (m.numTimesToErrOnCreateParty == 0) && m.errOnCreateParty {
		return 0, errors.New("error")
	}
	return 1, nil
}

func (m *mockPartyStore) GetPartyByShortID(ctx context.Context, shortID string) (store.Party, error) {
	return store.Party{}, m.errGetPartyByShortID
}

func (m *mockPartyStore) CreatePartyMember(ctx context.Context, idMember, idParty int) error {
	return m.errCreatePartyMember
}
