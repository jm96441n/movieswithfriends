package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jm96441n/movieswithfriends/store"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPartyMembersQueries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	partyMemberDBContainer := setupDBContainer(ctx, t)

	testCases := map[string]func(t *testing.T){
		"testCreatePartyMember":                            testCreatePartyMember(ctx, partyMemberDBContainer),
		"testCreatePartyMemberFailsOnDuplicatePartyMember": testCreatePartyMemberFailsOnDuplicatePartyMember(ctx, partyMemberDBContainer),
	}

	for name, tc := range testCases {
		t.Run(name, tc)
	}
}

func testCreatePartyMember(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

		idMember, idParty := seedCreatePartyMemberBackground(t, testConn)

		err := db.CreatePartyMember(context.Background(), idMember, idParty)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testCreatePartyMemberFailsOnDuplicatePartyMember(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

		idMember, idParty := seedCreatePartyMemberBackground(t, testConn)

		err := db.CreatePartyMember(context.Background(), idMember, idParty)
		if err != nil {
			t.Error(err)
			return
		}

		err = db.CreatePartyMember(context.Background(), idMember, idParty)
		if !errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
			t.Errorf("expected %v, got %v", store.ErrMemberPartyCombinationNotUnique, err)
		}
	}
}

func seedCreatePartyMemberBackground(t *testing.T, conn *pgx.Conn) (int, int) {
	t.Helper()
	var (
		idMember int
		idParty  int
	)
	err := conn.QueryRow(context.Background(), "insert into parties (name, short_id) values($1, $2) returning id_party", "test-party", "abcdef").Scan(&idParty)
	if err != nil {
		t.Fatal(err)
	}

	err = conn.QueryRow(context.Background(), "insert into profiles (first_name, last_name) values($1, $2) returning id_profile", "tom", "bomba").Scan(&idMember)
	if err != nil {
		t.Fatal(err)
	}
	return idMember, idParty
}
