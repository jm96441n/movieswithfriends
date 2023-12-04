package store

import (
	"context"
)

type Account struct {
	IDAccount    int
	Login        string
	PasswordHash string
	Profile      Profile
}

type Party struct {
	IDParty int
	Name    string
}

type Profile struct {
	IDProfile int `db:"id_profile"`
	Name      string
	AccountID int
	Parties   []Party
}

const GetProfileQuery = `SELECT profiles.id_profile, profiles.name, parties.id_party, parties.name  FROM profiles
JOIN accounts on accounts.id_account = profiles.id_account
JOIN profile_parties on profile_parties.id_profile = profiles.id_profile
JOIN parties on profile_parties.id_party = parties.id_party
WHERE accounts.login = $1;`

func (pg *PGStore) GetProfile(ctx context.Context, login string) (Profile, error) {
	rows, err := pg.db.Query(ctx, GetProfileQuery, login)
	if err != nil {
		return Profile{}, err
	}

	profile := Profile{Parties: make([]Party, 0)}
	for rows.Next() {
		var party Party
		rows.Scan(&profile.IDProfile, &profile.Name, &party.IDParty, &party.Name)
		profile.Parties = append(profile.Parties, party)
	}

	return profile, nil
}

const (
	CreateAccountQuery     = `INSERT INTO accounts(login, password_hash) VALUES ($1, $2) RETURNING "id_account"`
	CreateProfileQuery     = `INSERT INTO profiles(name, id_account) VALUES ($1, $2) RETURNING "id_profile"`
	GetAccountByLoginQuery = `SELECT id_account, login, password_hash FROM accounts WHERE accounts.login = $1`
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
	account := Account{}
	// profile := Profile{}
	row.Scan(&account.IDAccount, &account.Login, &account.PasswordHash)
	return account, nil
}
