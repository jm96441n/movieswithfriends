package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Account struct {
	ID           int
	Login        string
	PasswordHash string
	Profile      Profile
}

type Profile struct {
	ID        int
	Name      string
	AccountID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const GetProfileQuery = `SELECT * FROM profiles
WHERE profiles.id = $1`

func (pg *PGStore) GetProfile(ctx context.Context, id int) (Profile, error) {
	rows, err := pg.db.Query(ctx, GetProfileQuery, id)
	if err != nil {
		return Profile{}, err
	}
	profile, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Profile])
	if err != nil {
		return Profile{}, err
	}

	return profile, nil
}

const (
	CreateAccountQuery     = `INSERT INTO accounts(login, password_hash) VALUES ($1, $2) RETURNING "id_account"`
	CreateProfileQuery     = `INSERT INTO profiles(name, id_account) VALUES ($1, $2) RETURNING "id_profile"`
	GetAccountByLoginQuery = `SELECT * FROM accounts JOIN profiles ON profiles.account_id = accounts.id WHERE accounts.login = $1`
)

func (pg *PGStore) CreateAccount(ctx context.Context, name, login string, passwordHash []byte) error {
	tx, err := pg.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	var id_account int

	row := tx.QueryRow(ctx, CreateAccountQuery, login, passwordHash)
	err = row.Scan(&id_account)
	if err != nil {
		return err
	}

	tx.Exec(ctx, CreateProfileQuery, name, id_account)

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PGStore) GetAccountByLogin(ctx context.Context, login string) (Account, error) {
	row := pg.db.QueryRow(ctx, GetAccountByLoginQuery, login)
	fmt.Println(row)
	account := Account{}
	profile := Profile{}
	row.Scan(&account)
	row.Scan(&profile)
	return account, nil
}