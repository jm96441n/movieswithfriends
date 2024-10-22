package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

type Authenticator struct {
	Logger            *slog.Logger
	AccountRepository accountRepository
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (a *Authenticator) Authenticate(ctx context.Context, email, password string) (store.Account, error) {
	account, err := a.AccountRepository.FindAccountByEmail(ctx, email)
	if err != nil {
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		a.Logger.Error("error hashing password", slog.Any("error", err))
		return store.Account{}, err
	}
	account, err := a.AccountRepository.CreateAccount(ctx, req.Email, req.FirstName, req.LastName, hashedPassword)
	if err != nil {
		a.Logger.Error("error creating account", slog.Any("error", err))
		return store.Account{}, err
	}
	return account, nil
}
