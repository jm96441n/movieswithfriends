package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgUniqueViolationCode = "23505"

type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
	Profile  Profile
}

const insertAccountQuery = `insert into accounts (email, password) values ($1, $2) returning id_account`

var (
	ErrDuplicateEmailAddress = errors.New("email address already exists")
	ErrNotFound              = errors.New("not found")
)

func (p *ProfileRepository) CreateAccount(ctx context.Context, email, firstName, lastName string, password []byte) (Account, error) {
	txn, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Account{}, err
	}

	defer txn.Rollback(ctx)

	account := Account{Email: email}

	err = txn.QueryRow(ctx, insertAccountQuery, email, password).Scan(&account.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgUniqueViolationCode {
				if pgErr.ConstraintName == "unique_email_addresses" {
					return Account{}, ErrDuplicateEmailAddress
				}
			}
		}
		return Account{}, err
	}

	profile, err := p.CreateProfileWithTxn(ctx, txn, firstName, lastName, account.ID)
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

const findAccountByEmail = `
select accounts.id_account, accounts.email, accounts.password, profiles.id_profile from accounts 
join profiles on profiles.id_account = accounts.id_account
where accounts.email = $1
`

func (p *ProfileRepository) FindAccountByEmail(ctx context.Context, email string) (Account, error) {
	var account Account
	err := p.db.QueryRow(ctx, findAccountByEmail, email).Scan(&account.ID, &account.Email, &account.Password, &account.Profile.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	return account, nil
}

const accountExistsQuery = `SELECT EXISTS(select true from accounts where id_account = $1)`

func (p *ProfileRepository) AccountExists(ctx context.Context, id int) (bool, error) {
	var exists bool

	err := p.db.QueryRow(ctx, accountExistsQuery, id).Scan(&exists)

	return exists, err
}

const getAccountAndProfileInfoQuery = `
select accounts.email, profiles.first_name, profiles.last_name from accounts
join profiles on profiles.id_account = accounts.id_account
where accounts.id_account = $1
`

func (p *ProfileRepository) GetAccountAndProfileInfo(ctx context.Context, id int) (Account, error) {
	var account Account
	err := p.db.QueryRow(ctx, getAccountAndProfileInfoQuery, id).Scan(&account.Email, &account.Profile.FirstName, &account.Profile.LastName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Account{}, ErrNotFound
		}
		return Account{}, err
	}
	return account, nil
}

type AccountUpdateAttrs struct {
	ID       int
	Email    string
	Password []byte
}

type ProfileUpdateAttrs struct {
	ID        int
	FirstName string
	LastName  string
}

func (p *ProfileRepository) UpdateProfileAndAccount(ctx context.Context, accountAttrs AccountUpdateAttrs, profileAttrs ProfileUpdateAttrs) error {
	txn, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer txn.Rollback(ctx)

	err = updateAccount(ctx, txn, accountAttrs)
	if err != nil {
		return err
	}

	err = updateProfile(ctx, txn, profileAttrs)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func updateAccount(ctx context.Context, txn pgx.Tx, attrs AccountUpdateAttrs) error {
	if len(attrs.Password) == 0 {
		fmt.Println(attrs.Email)
		_, err := txn.Exec(ctx, `update accounts set email = $1 where id_account = $2`, attrs.Email, attrs.ID)
		if err != nil {
			return err
		}
		return nil
	}

	_, err := txn.Exec(ctx, `update accounts set email = $1, password = $2 where id_account = $3`, attrs.Email, attrs.Password, attrs.ID)
	if err != nil {
		return err
	}

	return nil
}

func updateProfile(ctx context.Context, txn pgx.Tx, attrs ProfileUpdateAttrs) error {
	_, err := txn.Exec(ctx, `update profiles set first_name = $1, last_name = $2 where id_profile = $3`, attrs.FirstName, attrs.LastName, attrs.ID)
	if err != nil {
		return err
	}

	return nil
}
