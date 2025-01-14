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

func TestGetPartyByID(t *testing.T) {
	ctx := context.Background()

	t.Parallel()
	schemaName := fmt.Sprintf("%s_get_party_by_id_schema", baseSchemaName)
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	repo := store.NewPartyRepository(connPool)
	idParty := seedParty(ctx, t, connPool, "test-party", "abcdef")

	testCases := map[string]struct {
		idParty              int
		expectedErr          error
		expectedPartyName    string
		expectedPartyShortID string
	}{
		"partyExists": {
			idParty:              idParty,
			expectedErr:          nil,
			expectedPartyName:    "test-party",
			expectedPartyShortID: "abcdef",
		},
		"partyDoesNotExist": {
			idParty:              0,
			expectedErr:          store.ErrNoRecord,
			expectedPartyName:    "",
			expectedPartyShortID: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			res, err := repo.GetPartyByID(ctx, tc.idParty)
			testhelpers.Assert(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)
			testhelpers.Assert(t, tc.expectedPartyName == res.Name, "expected %v, got %v", tc.expectedPartyName, res.Name)
			testhelpers.Assert(t, tc.expectedPartyShortID == res.ShortID, "expected %v, got %v", tc.expectedPartyShortID, res.ShortID)
			testhelpers.Assert(t, tc.idParty == res.ID, "expected %v, got %v", tc.idParty, res.ID)
		})
	}
}

func TestCreatePartySuccess(t *testing.T) {
	ctx := context.Background()

	t.Parallel()
	schemaName := fmt.Sprintf("%s_create_party_success_schema", baseSchemaName)
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	repo := store.NewPartyRepository(connPool)

	// create profile
	idMember := seedProfile(ctx, t, connPool)

	originalPartyCount := getPartyCount(ctx, t, connPool)

	partyID, err := repo.CreateParty(context.Background(), idMember, "test-party", "abcdef")

	// ensure no error from creation
	testhelpers.Ok(t, err, "expected error to be nil, got %v", err)

	// ensure party was persited
	got := getPartyCount(ctx, t, connPool)
	want := originalPartyCount + 1
	testhelpers.Assert(t, got == want, "expected %v, got %v", want, got)

	// ensure user is owner of party
	idOwner := getOwnerForParty(ctx, t, connPool, partyID)
	testhelpers.Assert(t, idOwner == idMember, "expected %v, got %v", idMember, idOwner)

	// check values are correct
	gotPartyName, gotPartyShortID := getParty(ctx, t, connPool, partyID)
	testhelpers.Assert(t, gotPartyName == "test-party", "expected %v, got %v", "test-party", gotPartyName)
	testhelpers.Assert(t, gotPartyShortID == "abcdef", "expected %v, got %v", "abcdef", gotPartyShortID)
}

func TestCreatePartyFailsDueToDuplicateShortID(t *testing.T) {
	ctx := context.Background()

	t.Parallel()
	schemaName := fmt.Sprintf("%s_create_party_fail_dupe_short_id_schema", baseSchemaName)
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	repo := store.NewPartyRepository(connPool)

	// create profile
	idMember := seedProfile(ctx, t, connPool)
	seedParty(ctx, t, connPool, "another-one", "abcdef")

	originalPartyCount := getPartyCount(ctx, t, connPool)

	_, err := repo.CreateParty(context.Background(), idMember, "test-party", "abcdef")

	// ensure no error from creation
	testhelpers.Assert(t, errors.Is(err, store.ErrDuplicatePartyShortID), "expected error to be %v, got %v", store.ErrDuplicatePartyShortID, err)

	// ensure party was not persited
	got := getPartyCount(ctx, t, connPool)
	want := originalPartyCount
	testhelpers.Assert(t, got == want, "expected %v, got %v", want, got)
}

func TestCreatePartyMember(t *testing.T) {
	ctx := context.Background()
	testCases := map[string]struct {
		expectedErr                      error
		backgroundSeedFn                 func(context.Context, *testing.T, *pgxpool.Pool, int, int)
		expectedPartyMemberCountIncrease int
	}{
		"successAddMember": {
			expectedErr:                      nil,
			backgroundSeedFn:                 func(ctx context.Context, t *testing.T, conn *pgxpool.Pool, id int, partyID int) {},
			expectedPartyMemberCountIncrease: 1,
		},
		"memberAlreadyExistsInParty": {
			expectedErr: store.ErrMemberPartyCombinationNotUnique,
			backgroundSeedFn: func(ctx context.Context, t *testing.T, conn *pgxpool.Pool, id int, partyID int) {
				query := `insert into party_members (id_member, id_party) values($1, $2)`
				_, err := conn.Exec(ctx, query, id, partyID)
				testhelpers.Ok(t, err, "failed to insert party member")
			},
			expectedPartyMemberCountIncrease: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			schemaName := fmt.Sprintf("%s_create_party_member_%s_schema", baseSchemaName, name)
			connPool := testhelpers.SetupConnPool(ctx, t, schemaName)

			t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

			// create party and profile
			idParty := seedParty(ctx, t, connPool, "test-party", "abcdef")
			idMember := seedProfile(ctx, t, connPool)

			// run the test specific seeding
			tc.backgroundSeedFn(ctx, t, connPool, idMember, idParty)

			originalPartyMemberCount := getPartyMemberCount(ctx, t, connPool, idParty)

			repo := store.NewPartyRepository(connPool)

			err := repo.CreatePartyMember(context.Background(), idMember, idParty)
			// ensure we got the expected error
			testhelpers.Assert(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)

			// check that the party member count increased
			got := getPartyMemberCount(ctx, t, connPool, idParty)
			want := originalPartyMemberCount + tc.expectedPartyMemberCountIncrease
			testhelpers.Assert(t, got == want, "expected %v, got %v", want, got)
		})
	}
}

func seedParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, name, shortID string) int {
	t.Helper()
	var idParty int
	err := conn.QueryRow(ctx, "insert into parties (name, short_id) values($1, $2) returning id_party", name, shortID).Scan(&idParty)
	testhelpers.Ok(t, err, "failed to insert party")

	return idParty
}

func seedProfile(ctx context.Context, t *testing.T, conn *pgxpool.Pool) int {
	var idMember int

	err := conn.QueryRow(ctx, "insert into profiles (first_name, last_name) values($1, $2) returning id_profile", "tom", "bomba").Scan(&idMember)

	testhelpers.Ok(t, err, "failed to insert profile")
	return idMember
}

func getPartyMemberCount(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID int) int {
	t.Helper()
	var count int
	err := conn.QueryRow(ctx, "select count(*) from party_members where id_party = $1", partyID).Scan(&count)
	testhelpers.Ok(t, err, "failed to get party member count")
	return count
}

func getPartyCount(ctx context.Context, t *testing.T, conn *pgxpool.Pool) int {
	t.Helper()
	var count int
	err := conn.QueryRow(ctx, "select count(*) from parties").Scan(&count)
	testhelpers.Ok(t, err, "failed to get party count")
	return count
}

func getOwnerForParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID int) int {
	t.Helper()
	var idOwner int
	err := conn.QueryRow(ctx, "select id_member from party_members where id_party = $1 AND owner = true", partyID).Scan(&idOwner)
	testhelpers.Ok(t, err, "failed to get owner for party")
	return idOwner
}

func getParty(ctx context.Context, t *testing.T, conn *pgxpool.Pool, partyID int) (string, string) {
	t.Helper()
	var name, shortID string
	err := conn.QueryRow(ctx, "select name, short_id from parties where id_party = $1", partyID).Scan(&name, &shortID)
	testhelpers.Ok(t, err, "failed to get party")
	return name, shortID
}
