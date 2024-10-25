package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jm96441n/movieswithfriends/store"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestProfilesQueries(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	profilesDBContainer := setupDBContainer(ctx, t)

	testCases := map[string]func(t *testing.T){
		"testGetProfileByID":                              testGetProfileByID(ctx, profilesDBContainer),
		"testGetProfileByIDErrorsWhenProfileDoesNotExist": testGetProfileByIDErrorsWhenProfileDoesNotExist(ctx, profilesDBContainer),
		"testCreateProfileWithTxn":                        testCreateProfileWithTxn(ctx, profilesDBContainer),
	}

	for name, tc := range testCases {
		t.Run(name, tc)
	}
}

func testGetProfileByID(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

		account := seedAccount(ctx, t, testConn)

		expectedProfile := store.Profile{
			FirstName: "John",
			LastName:  "Doe",
			AccountID: account.ID,
		}

		idProfile := seedProfile(ctx, t, testConn, expectedProfile)

		p, err := db.GetProfileByID(ctx, idProfile)
		if err != nil {
			t.Error(err)
		}

		if p.FirstName != expectedProfile.FirstName {
			t.Errorf("expected first name %s, got %s", expectedProfile.FirstName, p.FirstName)
		}

		if p.LastName != expectedProfile.LastName {
			t.Errorf("expected last name %s, got %s", expectedProfile.LastName, p.LastName)
		}
	}
}

func testGetProfileByIDErrorsWhenProfileDoesNotExist(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

		idProfile := 1

		_, err := db.GetProfileByID(ctx, idProfile)
		if !errors.Is(err, store.ErrNoRecord) {
			t.Error(err)
		}
	}
}

func testCreateProfileWithTxn(ctx context.Context, ctr *postgres.PostgresContainer) func(t *testing.T) {
	return func(t *testing.T) {
		db, testConn := setupDB(ctx, t, ctr)
		t.Cleanup(cleanupAndResetDB(ctx, t, ctr, testConn, db))

		txn, err := testConn.Begin(ctx)
		if err != nil {
			t.Error(err)
			return
		}

		account := seedAccount(ctx, t, testConn)

		expectedProfile := store.Profile{
			FirstName: "John",
			LastName:  "Doe",
		}

		p, err := db.CreateProfileWithTxn(ctx, txn, expectedProfile.FirstName, expectedProfile.LastName, account.ID)
		if err != nil {
			t.Error(err)
			return
		}

		err = txn.Commit(ctx)
		if err != nil {
			t.Error(err)
			return
		}

		actualProfile, err := db.GetProfileByID(ctx, p.ID)
		if err != nil {
			t.Error(err)
			return
		}

		if actualProfile.FirstName != expectedProfile.FirstName {
			t.Errorf("expected first name %s, got %s", expectedProfile.FirstName, actualProfile.FirstName)
		}

		if actualProfile.LastName != expectedProfile.LastName {
			t.Errorf("expected last name %s, got %s", expectedProfile.LastName, actualProfile.LastName)
		}

		if actualProfile.AccountID != account.ID {
			t.Errorf("expected account id %d, got %d", account.ID, actualProfile.AccountID)
		}
	}
}

func seedProfile(ctx context.Context, t *testing.T, testConn *pgx.Conn, p store.Profile) int {
	t.Helper()
	var id int
	err := testConn.QueryRow(ctx, "insert into profiles (first_name, last_name, id_account) values($1, $2, $3) returning id_profile", p.FirstName, p.LastName, p.AccountID).Scan(&id)
	if err != nil {
		t.Error(err)
		return 0
	}

	return id
}

func seedAccount(ctx context.Context, t *testing.T, testConn *pgx.Conn) store.Account {
	t.Helper()

	account := store.Account{
		Email:    "email@email.com",
		Password: []byte("unhashed password"),
	}
	err := testConn.QueryRow(ctx, "insert into accounts (email, password) values($1, $2) returning id_account", account.Email, account.Password).Scan(&account.ID)
	if err != nil {
		t.Error(err)
		return store.Account{}
	}
	return account
}
