package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func SeedAccountWithProfile(ctx context.Context, t *testing.T, conn *pgxpool.Pool, email, password, firstName, lastName string) {
	t.Helper()

	txn, err := conn.Begin(ctx)
	Ok(t, err, "Failed to open transaction to seed account")

	defer txn.Rollback(ctx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	Ok(t, err, "failed to hash password")

	var accountID int
	err = txn.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) returning id_account", email, hashedPassword).Scan(&accountID)
	Ok(t, err, "failed to insert account")

	_, err = txn.Exec(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3)", firstName, lastName, accountID)
	Ok(t, err, "failed to insert profile")

	Ok(t, txn.Commit(ctx), "failed to commit transaction")
}
