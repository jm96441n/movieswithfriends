package identityaccess_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/identityaccess"
	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/testhelpers"
	"golang.org/x/crypto/bcrypt"
)

func TestSignupReq_Validate(t *testing.T) {
	testCases := map[string]struct {
		req  *identityaccess.SignupReq
		want error
	}{
		"validRequest": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "1Password",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: nil,
		},
		"missingEmail": {
			req: &identityaccess.SignupReq{
				Email:     "",
				Password:  "1Password",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: &identityaccess.SignupValidationError{EmailError: identityaccess.ErrEmptyEmail},
		},
		"missingPassword": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: &identityaccess.SignupValidationError{PasswordError: identityaccess.ErrPasswordTooShort},
		},
		"passwordMissingNumber": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "Password",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: &identityaccess.SignupValidationError{PasswordError: identityaccess.ErrPasswordMissingNumber},
		},
		"passwordMissingUppercase": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "1password",
				FirstName: "FirstName",
				LastName:  "Lastname",
			},
			want: &identityaccess.SignupValidationError{PasswordError: identityaccess.ErrPasswordMissingUppercaseChar},
		},
		"missingFirstName": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "1Password",
				FirstName: "",
				LastName:  "Lastname",
			},
			want: &identityaccess.SignupValidationError{FirstNameError: identityaccess.ErrMissingFirstName},
		},
		"missingLastName": {
			req: &identityaccess.SignupReq{
				Email:     "email@email.com",
				Password:  "1Password",
				FirstName: "FirstName",
				LastName:  "",
			},
			want: &identityaccess.SignupValidationError{LastNameError: identityaccess.ErrMissingLastName},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			err := tc.req.Validate()

			testhelpers.Assert(t, errors.Is(err, tc.want), "expected %v, got %v", tc.want, err)

			if err != nil {
				got := &identityaccess.SignupValidationError{}
				want := &identityaccess.SignupValidationError{}
				testhelpers.Assert(t, errors.As(err, &got), "expected %v, got %v", tc.want, got)
				testhelpers.Assert(t, errors.As(tc.want, &want), "expected %v, got %v", tc.want, want)

				testhelpers.Assert(t, errors.Is(got.EmailError, want.EmailError), "expected %v, got %v", want.EmailError, got.EmailError)
				testhelpers.Assert(t, errors.Is(got.PasswordError, want.PasswordError), "expected %v, got %v", want.PasswordError, got.PasswordError)
				testhelpers.Assert(t, errors.Is(got.FirstNameError, want.FirstNameError), "expected %v, got %v", want.FirstNameError, got.FirstNameError)
				testhelpers.Assert(t, errors.Is(got.LastNameError, want.LastNameError), "expected %v, got %v", want.LastNameError, got.LastNameError)
			}
		})
	}
}

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaName := "identityaccess_authenticator_authenticate_schema"
	connPool := SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })
	existingUserEmail := "email@email.com"
	existingUserPassword := "1Password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(existingUserPassword), bcrypt.DefaultCost)

	testhelpers.Ok(t, err, "failed to generate password hash")

	seedProfile(ctx, t, connPool, existingUserEmail, hashedPassword)

	authenticator := &identityaccess.Authenticator{
		ProfileRepository: store.NewProfileRepository(connPool),
		Logger:            slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

	testCases := map[string]struct {
		expectedErr error
		email       string
		password    string
	}{
		"accountIsAuthenticated": {
			expectedErr: nil,
			email:       existingUserEmail,
			password:    existingUserPassword,
		},
		"emailDoesNotExist": {
			expectedErr: identityaccess.ErrInvalidCredentials,
			email:       "none@email.com",
			password:    existingUserPassword,
		},
		"passwordIsIncorrect": {
			expectedErr: identityaccess.ErrInvalidCredentials,
			email:       existingUserEmail,
			password:    "wrongPassword",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			_, err := authenticator.Authenticate(ctx, tc.email, tc.password)
			testhelpers.Assert(t, errors.Is(err, tc.expectedErr), "expected %v, got %v", tc.expectedErr, err)
		})
	}
}

func TestAccountExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	schemaName := "identityaccess_authenticator_account_exists_schema"
	connPool := SetupConnPool(ctx, t, schemaName)

	t.Cleanup(func() { testhelpers.CleanupAndResetDB(ctx, t, connPool, schemaName) })
	existingUserEmail := "email@email.com"
	existingUserPassword := "1Password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(existingUserPassword), bcrypt.DefaultCost)

	testhelpers.Ok(t, err, "failed to generate password hash")

	existingAccountID := seedProfile(ctx, t, connPool, existingUserEmail, hashedPassword)

	authenticator := &identityaccess.Authenticator{
		ProfileRepository: store.NewProfileRepository(connPool),
		Logger:            slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}

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

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			got, err := authenticator.AccountExists(ctx, tc.accountID)
			testhelpers.Ok(t, err, "expected not error, got %v", err)
			testhelpers.Assert(t, got == tc.want, "expected %v, got %v", tc.want, got)
		})
	}
}

func seedProfile(ctx context.Context, t *testing.T, connPool *pgxpool.Pool, email string, hashedPassword []byte) int {
	t.Helper()

	var accountID int
	err := connPool.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) RETURNING id_account", email, hashedPassword).Scan(&accountID)
	testhelpers.Ok(t, err, "failed to insert account")

	_, err = connPool.Exec(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3)", "name", "name", accountID)
	testhelpers.Ok(t, err, "failed to insert account")

	return accountID
}
