package store

import "context"

type Account struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte
	Profile  Profile
}

const insertAccountQuery = `insert into accounts (email, password) values ($1, $2, $3) returning id_account`

func (pg *PGStore) CreateAccount(ctx context.Context, email, firstName, lastName string, password []byte) (Account, error) {
	return Account{}, nil
}
