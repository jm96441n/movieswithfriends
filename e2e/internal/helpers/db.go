package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type TestAccountInfo struct {
	AccountID int
	ProfileID int
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func SeedAccountWithProfile(ctx context.Context, t *testing.T, conn *pgxpool.Pool, accountInfo TestAccountInfo) TestAccountInfo {
	t.Helper()

	txn, err := conn.Begin(ctx)
	Ok(t, err, "Failed to open transaction to seed account")

	defer txn.Rollback(ctx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(accountInfo.Password), bcrypt.DefaultCost)
	Ok(t, err, "failed to hash password")

	err = txn.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) returning id_account", accountInfo.Email, hashedPassword).Scan(&accountInfo.AccountID)
	Ok(t, err, "failed to insert account")

	err = txn.QueryRow(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3) returning id_profile", accountInfo.FirstName, accountInfo.LastName, accountInfo.AccountID).Scan(&accountInfo.ProfileID)
	Ok(t, err, "failed to insert profile")

	Ok(t, txn.Commit(ctx), "failed to commit transaction")

	return accountInfo
}
