package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
	"github.com/jm96441n/movieswithfriends/testhelpers"
)

const baseSchemaName = "partymgmt"

func TestCreatePartyMember(t *testing.T) {
	ctx := context.Background()
	testCases := map[string]struct {
		expectedErr error
	}{
		"successAddMember": {
			expectedErr: nil,
		},
		// TODO:: seed this correctly
		"memberAlreadyExistsInParty": {
			expectedErr: store.ErrMemberPartyCombinationNotUnique,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			schemaName := fmt.Sprintf("%s_create_party_member_%s_schema", baseSchemaName, name)
			connPool := SetupConnPool(ctx, t, schemaName)

			t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

			idMember, idParty := seedPartyAndProfile(ctx, t, connPool)
			repo := store.NewPartyRepository(connPool)

			err := repo.CreatePartyMember(context.Background(), idMember, idParty)
			testhelpers.Assert(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)
		})
	}
}

// func testCreatePartyMemberFailsOnDuplicatePartyMember(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
// 	return func(t *testing.T) {
// 		db, testConn := setupDB(ctx, t, ctr)
// 		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))
//
// 		idMember, idParty := seedCreatePartyMemberBackground(t, testConn)
//
// 		err := db.CreatePartyMember(context.Background(), idMember, idParty)
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}
//
// 		err = db.CreatePartyMember(context.Background(), idMember, idParty)
// 		if !errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
// 			t.Errorf("expected %v, got %v", store.ErrMemberPartyCombinationNotUnique, err)
// 		}
// 	}
// }

func seedPartyAndProfile(ctx context.Context, t *testing.T, conn *pgxpool.Pool) (int, int) {
	t.Helper()
	var (
		idMember int
		idParty  int
	)
	err := conn.QueryRow(ctx, "insert into parties (name, short_id) values($1, $2) returning id_party", "test-party", "abcdef").Scan(&idParty)
	testhelpers.Ok(t, err, "failed to insert party")

	err = conn.QueryRow(ctx, "insert into profiles (first_name, last_name) values($1, $2) returning id_profile", "tom", "bomba").Scan(&idMember)
	testhelpers.Ok(t, err, "failed to insert profile")
	return idMember, idParty
}
