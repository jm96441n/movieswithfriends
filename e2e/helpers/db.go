package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func SeedAccountWithProfile(ctx context.Context, t *testing.T, conn *pgx.Conn, email, password, firstName, lastName string) {
	t.Helper()

	txn, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	defer txn.Rollback(ctx)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	var accountID int
	err = txn.QueryRow(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2) returning id_account", email, hashedPassword).Scan(&accountID)
	if err != nil {
		t.Fatalf("failed to insert account to db: %v", err)
	}

	_, err = txn.Exec(ctx, "INSERT INTO profiles (first_name, last_name, id_account) VALUES ($1, $2, $3)", firstName, lastName, accountID)
	if err != nil {
		t.Fatalf("failed to insert profile to db: %v", err)
	}

	txn.Commit(ctx)
}
