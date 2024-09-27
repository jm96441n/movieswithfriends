package store

import (
	"context"
	"errors"

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

const findAccountByEmail = `select id_account, email, password from accounts where email = $1`

func (pg *PGStore) FindAccountByEmail(ctx context.Context, email string) (Account, error) {
	var account Account
	err := pg.db.QueryRow(ctx, findAccountByEmail, email).Scan(&account.ID, &account.Email, &account.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	return account, nil
}
