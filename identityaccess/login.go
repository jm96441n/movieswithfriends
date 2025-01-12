package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"golang.org/x/crypto/bcrypt"
)

type SignupReq struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	PartyID   string `json:"partyID"`
}

type SignupValidationError struct {
	EmailError     error
	PasswordError  error
	FirstNameError error
	LastNameError  error
}

func (s *SignupValidationError) Error() string {
	return fmt.Sprintf("signup validation error: %#v", s)
}

func (s *SignupValidationError) IsNil() bool {
	return s.EmailError == nil && s.PasswordError == nil && s.FirstNameError == nil && s.LastNameError == nil
}

var (
	ErrEmptyEmail                   = errors.New("email is required")
	ErrPasswordTooShort             = errors.New("password must be at least 8 characters long")
	ErrPasswordMissingNumber        = errors.New("password must contain at least one number")
	ErrPasswordMissingUppercaseChar = errors.New("password must contain at least one uppercase character")
	ErrMissingFirstName             = errors.New("first name is required")
	ErrMissingLastName              = errors.New("last name is required")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountExists      = errors.New("account already exists")
)

var numRegex = regexp.MustCompile("[0-9]+")

func (s SignupReq) Validate() error {
	var err SignupValidationError
	if s.Email == "" {
		err.EmailError = ErrEmptyEmail
	}
	err.PasswordError = validatePassword(s.Password)

	if s.FirstName == "" {
		err.FirstNameError = ErrMissingFirstName
	}

	if s.LastName == "" {
		err.LastNameError = ErrMissingLastName
	}

	if !err.IsNil() {
		return &err
	}

	return nil
}

func validatePassword(password string) error {
	var err error
	if len(password) < 8 {
		err = errors.Join(err, ErrPasswordTooShort)
	}

	if strings.ToLower(password) == password || strings.ToUpper(password) == password {
		err = errors.Join(err, ErrPasswordMissingUppercaseChar)
	}

	if len(numRegex.FindAllString(password, -1)) == 0 {
		err = errors.Join(err, ErrPasswordMissingNumber)
	}

	if err != nil {
		return err
	}

	return nil
}

type Authenticator struct {
	Logger            *slog.Logger
	ProfileRepository *store.ProfileRepository
}

func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (Profile, error) {
	res, err := a.ProfileRepository.GetProfileByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			a.Logger.Error("account not found", slog.String("email", email))
			return Profile{}, ErrInvalidCredentials
		}
		a.Logger.Error("error finding profile by email", slog.Any("error", err), slog.String("email", email))
		return Profile{}, err
	}

	profile := convertProfileByEmailToProfile(res)

	err = bcrypt.CompareHashAndPassword(profile.Account.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			a.Logger.Error("incorrect password", slog.Any("error", err))
			return Profile{}, fmt.Errorf("%w: %s", ErrInvalidCredentials, err)
		}
		a.Logger.Error("error comparing password", slog.Any("error", err))
		return Profile{}, err
	}

	return profile, nil
}

func convertProfileByEmailToProfile(res store.GetProfileByEmailResult) Profile {
	return Profile{
		ID: res.ProfileID,
		Account: Account{
			ID:       res.AccountID,
			Email:    res.AccountEmail,
			Password: res.AccountPassword,
		},
	}
}

func (a *Authenticator) AccountExists(ctx context.Context, accountID int) (bool, error) {
	found, err := a.ProfileRepository.AccountExists(ctx, accountID)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			return false, nil
		}
		return false, err
	}
	return found, nil
}
