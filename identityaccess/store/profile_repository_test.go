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
	ctx := context.Background()
	t.Parallel()
	schemaName := fmt.Sprintf("%s_get_profile_by_id_schema", baseSchemaName)
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })
	expectedFirstName := "FistName"
	expectedLastName := "LastName"
	expectedEmail := "email@email.com"
	expectedPassword := []byte("password")

	existingProfileID, existingAccountID, createdAt := seedProfile(ctx, t, connPool, expectedFirstName, expectedLastName, expectedEmail, []byte(expectedPassword))
	cases := map[string]struct {
		profileID   int
		expectedErr error
		want        store.GetProfileResult
	}{
		"profileExists": {
			profileID:   existingProfileID,
			expectedErr: nil,
			want: store.GetProfileResult{
				ID:              existingProfileID,
				FirstName:       expectedFirstName,
				LastName:        expectedLastName,
				CreatedAt:       createdAt,
				AccountEmail:    expectedEmail,
				AccountID:       existingAccountID,
				AccountPassword: expectedPassword,
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

			testhelpers.Assert(t, got.FirstName == testCase.want.FirstName, "expected %v, got %v", testCase.want.FirstName, got.FirstName)
			testhelpers.Assert(t, got.LastName == testCase.want.LastName, "expected %v, got %v", testCase.want.LastName, got.LastName)
			testhelpers.Assert(t, got.AccountEmail == testCase.want.AccountEmail, "expected %v, got %v", testCase.want.AccountEmail, got.AccountEmail)
			testhelpers.Assert(t, got.CreatedAt.Equal(testCase.want.CreatedAt), "expected %v, got %v", testCase.want.CreatedAt, got.CreatedAt)
			testhelpers.Assert(t, bytes.Equal(got.AccountPassword, testCase.want.AccountPassword), "expected %v, got %v", testCase.want.AccountPassword, got.AccountPassword)
		})
	}
}

func TestGetProfileByEmail(t *testing.T) {
	t.Parallel()
	schemaName := fmt.Sprintf("%s_get_profile_by_email_schema", baseSchemaName)
	ctx := context.Background()
	connPool := testhelpers.SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })

	existingProfileID, existingAccountID, _ := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com", []byte("password"))
	cases := map[string]struct {
		profileEmail string
		expectedErr  error
		want         store.GetProfileResult
	}{
		"profileExists": {
			profileEmail: "email@email.com",
			expectedErr:  nil,
			want: store.GetProfileResult{
				ID:              existingProfileID,
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

			resultEqual := func(g, w store.GetProfileResult) bool {
				return g.ID == w.ID &&
					g.AccountEmail == w.AccountEmail &&
					g.AccountID == w.AccountID &&
					bytes.Equal(g.AccountPassword, w.AccountPassword)
			}

			testhelpers.Assert(t, resultEqual(got, testCase.want), "expected %v, got %v", testCase.want, got)
		})
	}
}

func TestCreateProfile(t *testing.T) {
	ctx := context.Background()

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
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			schemaName := fmt.Sprintf("%s_create_profile_schema_%s", baseSchemaName, name)
			connPool := testhelpers.SetupConnPool(ctx, tt, schemaName)
			tt.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, tt, connPool, schemaName) })

			// seed existing profile
			_, _, _ = seedProfile(ctx, tt, connPool, "FirstName", "LastName", "email@email.com", []byte("password"))
			repo := store.NewProfileRepository(connPool)

			actual, err := repo.CreateProfile(ctx, testCase.newProfileEmail, "FirstName", "LastName", []byte("password"))

			testhelpers.Assert(tt, errors.Is(err, testCase.expectedErr), "expected error %v, got %v", testCase.expectedErr, err)

			// ensure we get some id > 0 back
			if err == nil {
				testhelpers.Assert(tt, actual.AccountID > 0, "expected AccountID > 0, got %v", actual.AccountID)
				testhelpers.Assert(tt, actual.ProfileID > 0, "expected ProfileID > 0, got %v", actual.ProfileID)
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

	_, existingAccountID, _ := seedProfile(ctx, t, connPool, "FirstName", "LastName", "email@email.com", []byte("password"))

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
	ctx := context.Background()

	originalFirstName := "FirstName"
	originalLastName := "LastName"
	originalEmail := "email@email.com"
	originalPassword := []byte("password")

	newFirstName := "newName"
	newLastName := "newLastName"
	newEmail := "new@email.com"
	newPassword := []byte("newpassword")

	dupeEmail := "DUPE@email.com"

	testCases := map[string]struct {
		accountUpdateAttrs store.AccountUpdateAttrs
		profileUpdateAttrs store.ProfileUpdateAttrs
		expectedErr        error
		want               store.GetProfileResult
	}{
		"updateSucceedsNoPasswordChange": {
			expectedErr: nil,
			accountUpdateAttrs: store.AccountUpdateAttrs{
				Email: newEmail,
			},
			profileUpdateAttrs: store.ProfileUpdateAttrs{
				FirstName: newFirstName,
				LastName:  newLastName,
			},
			want: store.GetProfileResult{
				FirstName:       newFirstName,
				LastName:        newLastName,
				AccountEmail:    newEmail,
				AccountPassword: originalPassword,
			},
		},
		"updateSucceedsWithPasswordChange": {
			expectedErr: nil,
			accountUpdateAttrs: store.AccountUpdateAttrs{
				Email:    newEmail,
				Password: newPassword,
			},
			profileUpdateAttrs: store.ProfileUpdateAttrs{
				FirstName: newFirstName,
				LastName:  newLastName,
			},
			want: store.GetProfileResult{
				FirstName:       newFirstName,
				LastName:        newLastName,
				AccountEmail:    newEmail,
				AccountPassword: newPassword,
			},
		},
		"updateFailsWithoutPasswordDupeEmail": {
			expectedErr: store.ErrDuplicateEmailAddress,
			accountUpdateAttrs: store.AccountUpdateAttrs{
				Email: dupeEmail,
			},
			profileUpdateAttrs: store.ProfileUpdateAttrs{
				FirstName: "newName",
				LastName:  "newLastName",
			},
			want: store.GetProfileResult{
				FirstName:       originalFirstName,
				LastName:        originalLastName,
				AccountEmail:    originalEmail,
				AccountPassword: originalPassword,
			},
		},
		"updateFailsWithPasswordDupeEmail": {
			expectedErr: store.ErrDuplicateEmailAddress,
			accountUpdateAttrs: store.AccountUpdateAttrs{
				Email:    dupeEmail,
				Password: []byte("newPassword"),
			},
			profileUpdateAttrs: store.ProfileUpdateAttrs{
				FirstName: "newName",
				LastName:  "newLastName",
			},
			want: store.GetProfileResult{
				FirstName:       originalFirstName,
				LastName:        originalLastName,
				AccountEmail:    originalEmail,
				AccountPassword: originalPassword,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			schemaName := fmt.Sprintf("%s_update_profile_schema_%s", baseSchemaName, name)
			connPool := testhelpers.SetupConnPool(ctx, tt, schemaName)
			tt.Cleanup(func() {
				testhelpers.CleanupAndResetDB(ctx, tt, connPool, schemaName)
				connPool.Close()
			})

			existingProfileID, existingAccountID, _ := seedProfile(ctx, tt, connPool, originalFirstName, originalLastName, originalEmail, originalPassword)

			// seed dupe for duplicate email check
			seedProfile(ctx, tt, connPool, "another", "name", dupeEmail, []byte("anotherpassword"))

			testCase.accountUpdateAttrs.ID = existingAccountID
			testCase.profileUpdateAttrs.ID = existingProfileID
			testCase.want.AccountID = existingAccountID

			repo := store.NewProfileRepository(connPool)

			err := repo.UpdateProfile(ctx, testCase.accountUpdateAttrs, testCase.profileUpdateAttrs)
			testhelpers.Assert(tt, errors.Is(err, testCase.expectedErr), "expected %v error, got %v", testCase.expectedErr, err)

			got, err := repo.GetProfileByID(ctx, existingProfileID)
			testhelpers.Ok(tt, err, "failed to get updated profile")

			testhelpers.Assert(tt, got.FirstName == testCase.want.FirstName, "expected %v, got %v", testCase.want.FirstName, got.FirstName)
			testhelpers.Assert(tt, got.LastName == testCase.want.LastName, "expected %v, got %v", testCase.want.LastName, got.LastName)
			testhelpers.Assert(tt, got.AccountEmail == testCase.want.AccountEmail, "expected %v, got %v", testCase.want.AccountEmail, got.AccountEmail)
			testhelpers.Assert(tt, got.AccountID == testCase.want.AccountID, "expected %v, got %v", testCase.want.AccountID, got.AccountID)
			testhelpers.Assert(tt, bytes.Equal(got.AccountPassword, testCase.want.AccountPassword), "expected %v, got %v", testCase.want.AccountPassword, got.AccountPassword)
		})
	}
}

func TestGetProfileStats(t *testing.T) {
	// TODO: Add tests
}

func seedProfile(ctx context.Context, t *testing.T, connPool *pgxpool.Pool, firstName, lastName, email string, password []byte) (int, int, time.Time) {
	t.Helper()
	var (
		accountID int
		profileID int
		createdAt time.Time
	)

	err := connPool.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) RETURNING id_account", email, password).Scan(&accountID)
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
