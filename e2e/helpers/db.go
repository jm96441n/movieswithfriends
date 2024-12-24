package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func SeedAccount(ctx context.Context, t *testing.T, conn *pgx.Conn, email, password string) {
	t.Helper()
	_, err := conn.Exec(ctx, "INSERT INTO accounts (email, password) VALUES ($1, $2)", email, password)
	if err != nil {
		t.Fatal(err)
	}
}
