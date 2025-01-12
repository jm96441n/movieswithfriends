package store_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/testhelpers"
)

const baseSchemaName = "profile_repository_test"

func TestGetProfileByID(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_get_profile_by_id_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	existingProfileID, existingAccountID, createdAt := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com")
	cases := map[string]struct {
		profileID   int
		expectedErr error
		want        store.GetProfileByIDResult
	}{
		"profileExists": {
			profileID:   existingProfileID,
			expectedErr: nil,
			want: store.GetProfileByIDResult{
				ID:           existingProfileID,
				FirstName:    "FirstName",
				LastName:     "LastName",
				CreatedAt:    createdAt,
				AccountEmail: "email@email.com",
				AccountID:    existingAccountID,
			},
		},
		"profileDoesNotExist": {
			profileID:   0,
			expectedErr: store.ErrNoRecord,
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			repo := store.NewProfileRepository(connPool)

			got, err := repo.GetProfileByID(ctx, testCase.profileID)

			testhelpers.Assert(t, errors.Is(err, testCase.expectedErr), "expected error %v, got %v", testCase.expectedErr, err)

			testhelpers.Assert(t, got == testCase.want, "expected %v, got %v", testCase.want, got)
		})
	}
}

func TestGetProfileByEmail(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_get_profile_by_email_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	existingProfileID, existingAccountID, _ := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com")
	cases := map[string]struct {
		profileEmail string
		expectedErr  error
		want         store.GetProfileByEmailResult
	}{
		"profileExists": {
			profileEmail: "email@email.com",
			expectedErr:  nil,
			want: store.GetProfileByEmailResult{
				ProfileID:       existingProfileID,
				AccountEmail:    "email@email.com",
				AccountID:       existingAccountID,
				AccountPassword: []byte("password"),
			},
		},
		"profileDoesNotExist": {
			profileEmail: "doesnotExist@none.com",
			expectedErr:  store.ErrNoRecord,
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			repo := store.NewProfileRepository(connPool)

			got, err := repo.GetProfileByEmail(ctx, testCase.profileEmail)

			testhelpers.Assert(t, errors.Is(err, testCase.expectedErr), "expected error %v, got %v", testCase.expectedErr, err)

			resultEqual := func(g, w store.GetProfileByEmailResult) bool {
				return g.ProfileID == w.ProfileID &&
					g.AccountEmail == w.AccountEmail &&
					g.AccountID == w.AccountID &&
					bytes.Equal(g.AccountPassword, w.AccountPassword)
			}

			testhelpers.Assert(t, resultEqual(got, testCase.want), "expected %v, got %v", testCase.want, got)
		})
	}
}

func TestCreateProfile(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_create_profile_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	_, _, _ = seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com")

	testCases := map[string]struct {
		expectedErr     error
		newProfileEmail string
	}{
		"sucessfullyCreateProfile": {
			expectedErr:     nil,
			newProfileEmail: "newEmail@email.com",
		},
		"duplicateEmailCausesError": {
			expectedErr:     store.ErrDuplicateEmailAddress,
			newProfileEmail: "email@email.com",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			repo := store.NewProfileRepository(connPool)

			actual, err := repo.CreateProfile(ctx, testCase.newProfileEmail, "FirstName", "LastName", []byte("password"))

			testhelpers.Assert(t, errors.Is(err, testCase.expectedErr), "expected error %v, got %v", testCase.expectedErr, err)

			// ensure we get some id > 0 back
			if err == nil {
				testhelpers.Assert(t, actual.AccountID > 0, "expected AccountID > 0, got %v", actual.AccountID)
				testhelpers.Assert(t, actual.ProfileID > 0, "expected ProfileID > 0, got %v", actual.ProfileID)
			}
		})
	}
}

func TestAccountExists(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_account_exists_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	_, existingAccountID, _ := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com")

	testCases := map[string]struct {
		want      bool
		accountID int
	}{
		"accountExists": {
			want:      true,
			accountID: existingAccountID,
		},
		"accountDoesNotExist": {
			want:      false,
			accountID: 0,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			repo := store.NewProfileRepository(connPool)

			got, err := repo.AccountExists(ctx, testCase.accountID)

			testhelpers.Ok(t, err, "failed to check if account exists")
			testhelpers.Assert(t, got == testCase.want, "expected %v, got %v", testCase.want, got)
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_update_profile_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)
	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	existingProfileID, existingAccountID, _ := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com")

	testCases := map[string]struct {
		want               error
		accountUpdateAttrs store.AccountUpdateAttrs
		profileUpdateAttrs store.ProfileUpdateAttrs
	}{
		"updateSucceedsNoPasswordChange": {
			want: nil,
			accountUpdateAttrs: store.AccountUpdateAttrs{
				ID:    existingAccountID,
				Email: "new@email.com",
			},
			profileUpdateAttrs: store.ProfileUpdateAttrs{
				ID:        existingProfileID,
				FirstName: "newName",
				LastName:  "newLastName",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			repo := store.NewProfileRepository(connPool)

			got := repo.UpdateProfile(ctx, testCase.accountUpdateAttrs, testCase.profileUpdateAttrs)

			testhelpers.Assert(t, errors.Is(got, testCase.want), "expected %v, got %v", testCase.want, got)
			// get the updated profile and compare the values
		})
	}
}

func seedProfile(ctx context.Context, t *testing.T, connPool *pgxpool.Pool, firstName, lastName, email string) (int, int, time.Time) {
	t.Helper()
	var (
		accountID int
		profileID int
		createdAt time.Time
	)

	err := connPool.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) RETURNING id_account", email, "password").Scan(&accountID)
	testhelpers.Ok(t, err, "failed to insert account")

	err = connPool.QueryRow(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3) RETURNING id_profile, created_at", firstName, lastName, accountID).Scan(&profileID, &createdAt)
	testhelpers.Ok(t, err, "failed to insert account")

	return profileID, accountID, createdAt
}

func getProfileCount(ctx context.Context, t *testing.T, connPool *pgxpool.Pool) int {
	t.Helper()
	var count int

	err := connPool.QueryRow(ctx, "SELECT COUNT(*) FROM profiles").Scan(&count)
	testhelpers.Ok(t, err, "failed to get profile count")

	return count
}
