package store

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
	Profile  Profile
}

const insertAccountQuery = `insert into accounts (email, password) values ($1, $2) returning id_account`

func (pg *PGStore) CreateAccount(ctx context.Context, email, firstName, lastName string, password []byte) (Account, error) {
	txn, err := pg.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Account{}, err
	}

	defer txn.Rollback(ctx)

	account := Account{Email: email}

	err = txn.QueryRow(ctx, insertAccountQuery, email, password).Scan(&account.ID)
	if err != nil {
		return Account{}, err
	}

	profile, err := pg.CreateProfileWithTxn(ctx, txn, firstName, lastName, account.ID)
	if err != nil {
		return Account{}, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return Account{}, err
	}

	account.Profile = profile

	return account, nil
}

var ErrNotFound = errors.New("not found")

const findAccountByEmail = `
select accounts.id_account, accounts.email, accounts.password, profiles.id_profile from accounts 
join profiles on profiles.id_account = accounts.id_account
where accounts.email = $1
`

func (pg *PGStore) FindAccountByEmail(ctx context.Context, email string) (Account, error) {
	var account Account
	err := pg.db.QueryRow(ctx, findAccountByEmail, email).Scan(&account.ID, &account.Email, &account.Password, &account.Profile.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	return account, nil
}

const accountExistsQuery = `SELECT EXISTS(select true from accounts where id_account = $1)`

func (pg *PGStore) AccountExists(ctx context.Context, id int) (bool, error) {
	var exists bool

	err := pg.db.QueryRow(ctx, accountExistsQuery, id).Scan(&exists)

	return exists, err
}

const getAccountAndProfileInfoQuery = `
select accounts.email, profiles.first_name, profiles.last_name from accounts
join profiles on profiles.id_account = accounts.id_account
where accounts.id_account = $1
`

func (pg *PGStore) GetAccountAndProfileInfo(ctx context.Context, id int) (Account, error) {
	var account Account
	pg.logger.Info("GetAccountAndProfileInfo", slog.Any("id", id))
	err := pg.db.QueryRow(ctx, getAccountAndProfileInfoQuery, id).Scan(&account.Email, &account.Profile.FirstName, &account.Profile.LastName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	return account, nil
}
