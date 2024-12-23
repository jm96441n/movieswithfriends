package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jm96441n/movieswithfriends/store"
	"golang.org/x/crypto/bcrypt"
)

type accountRepository interface {
	FindAccountByEmail(context.Context, string) (store.Account, error)
	CreateAccount(context.Context, string, string, string, []byte) (store.Account, error)
	AccountExists(context.Context, int) (bool, error)
}

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
	return fmt.Sprintf("signup validation error: %v", s)
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
	if len(s.Password) < 8 {
		errors.Join(err.PasswordError, ErrPasswordTooShort)
	}

	if strings.ToLower(s.Password) == s.Password || strings.ToUpper(s.Password) == s.Password {
		errors.Join(err.PasswordError, ErrPasswordMissingUppercaseChar)
	}

	if len(numRegex.FindAllString("abc123def987asdf", -1)) == 0 {
		errors.Join(err.PasswordError, ErrPasswordMissingNumber)
	}

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

type Authenticator struct {
	Logger            *slog.Logger
	AccountRepository accountRepository
}

func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (store.Account, error) {
	account, err := a.AccountRepository.FindAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return store.Account{}, ErrInvalidCredentials
		}
		a.Logger.Error("error finding account by email", slog.Any("error", err), slog.String("email", email))
		return store.Account{}, err
	}

	err = bcrypt.CompareHashAndPassword(account.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			a.Logger.Error("incorrect password", slog.Any("error", err))
			return store.Account{}, fmt.Errorf("%w: %s", ErrInvalidCredentials, err)
		}
		a.Logger.Error("error comparing password", slog.Any("error", err))
		return store.Account{}, err
	}

	return account, nil
}

func (a *Authenticator) AccountExists(ctx context.Context, accountID int) (bool, error) {
	found, err := a.AccountRepository.AccountExists(ctx, accountID)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			return false, nil
		}
		return false, err
	}
	return found, nil
}

func (a *Authenticator) CreateAccount(ctx context.Context, req SignupReq) (store.Account, error) {
	err := req.Validate()
	if err != nil {
		return store.Account{}, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		a.Logger.Error("error hashing password", slog.Any("error", err))
		return store.Account{}, err
	}
	account, err := a.AccountRepository.CreateAccount(ctx, req.Email, req.FirstName, req.LastName, hashedPassword)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateEmailAddress) {
			a.Logger.Debug("email exists for account")
			return store.Account{}, ErrAccountExists
		}
		a.Logger.Error("error creating account", slog.Any("error", err))
		return store.Account{}, err
	}
	return account, nil
}
