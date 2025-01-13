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
	svc := &partymgmt.PartyService{}

	idMember := 1
	err := svc.AddNewMemberToParty(context.Background(), idMember, "shortID")
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
			svc := &partymgmt.PartyService{}

			idMember := 1
			err := svc.AddNewMemberToParty(context.Background(), idMember, "shortID")
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
			svc := &partymgmt.PartyService{}
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
			svc := &partymgmt.PartyService{}
			idMember := 1
			name := "name"
			_, err := svc.CreateParty(context.Background(), idMember, name)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
