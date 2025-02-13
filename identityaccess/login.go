package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/metrics"
	"golang.org/x/crypto/bcrypt"
)

type SignupReq struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
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

func (s *SignupValidationError) Is(err error) bool {
	_, ok := err.(*SignupValidationError)
	return ok
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

func (s SignupReq) Validate(ctx context.Context) error {
	_, span, _ := metrics.SpanFromContext(ctx, "signupReq.Validate")
	defer span.End()
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
	ProfileRepository *store.ProfileRepository
}

func (a *Authenticator) Authenticate(ctx context.Context, logger *slog.Logger, email, password string) (*Profile, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "authenticator.Authenticate")
	defer span.End()
	res, err := a.ProfileRepository.GetProfileByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			logger.ErrorContext(ctx, "account not found", slog.String("email", email))
			return nil, ErrInvalidCredentials
		}
		logger.ErrorContext(ctx, "error finding profile by email", slog.Any("error", err), slog.String("email", email))
		return nil, err
	}

	profile := convertGetProfileResultToProfile(ctx, res)
	profile.db = a.ProfileRepository

	err = bcrypt.CompareHashAndPassword(profile.Account.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			logger.ErrorContext(ctx, "incorrect password", slog.Any("error", err))
			return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err)
		}
		logger.ErrorContext(ctx, "error comparing password", slog.Any("error", err))
		return nil, err
	}

	return profile, nil
}

func (a *Authenticator) AccountExists(ctx context.Context, accountID int) (bool, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "authenticator.AccountExists")
	defer span.End()
	found, err := a.ProfileRepository.AccountExists(ctx, accountID)
	if err != nil {
		return false, err
	}
	return found, nil
}
