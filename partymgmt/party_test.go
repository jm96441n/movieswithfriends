package partymgmt_test

// func TestAddFriendToParty_HappyPath(t *testing.T) {
// 	t.Parallel()
// 	svc := &partymgmt.PartyService{}
//
// 	idMember := 1
// 	err := svc.AddNewMemberToParty(context.Background(), idMember, "shortID")
// 	if err != nil {
// 		t.Errorf("expected nil, got %v", err)
// 	}
// }
//
// func TestAddFriendToParty_SadPath(t *testing.T) {
// 	errGetPartyByShortID := errors.New("error getting party by short id")
// 	errCreatePartyMember := errors.New("error creating party member")
//
// 	testCases := map[string]struct {
// 		errGetPartyByShortID error
// 		errCreatePartyMember error
// 		expectedError        error
// 	}{
// 		"db failed to get party by short id": {
// 			errGetPartyByShortID: errGetPartyByShortID,
// 			expectedError:        errGetPartyByShortID,
// 		},
// 		"db failed to create party member": {
// 			errCreatePartyMember: errCreatePartyMember,
// 			expectedError:        errCreatePartyMember,
// 		},
// 		"db failed to create party member because member exists in party": {
// 			errCreatePartyMember: store.ErrMemberPartyCombinationNotUnique,
// 			expectedError:        partymgmt.ErrMemberExistsInParty,
// 		},
// 	}
//
// 	for name, tc := range testCases {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()
// 			svc := &partymgmt.PartyService{}
//
// 			idMember := 1
// 			err := svc.AddNewMemberToParty(context.Background(), idMember, "shortID")
// 			if err == nil {
// 				t.Errorf("expected error, got nil")
// 			}
//
// 			if !errors.Is(err, tc.expectedError) {
// 				t.Errorf("expected error %v, got %v", tc.expectedError, err)
// 			}
// 		})
// 	}
// }
