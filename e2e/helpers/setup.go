package helpers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func Setup(t *testing.T, dbCtr *postgres.PostgresContainer, page playwright.Page) (context.Context, *pgx.Conn) {
	t.Helper()
	ctx := context.Background()
	testConn := SetupDBConn(ctx, t, dbCtr)

	t.Cleanup(CleanupAndResetDB(ctx, t, dbCtr, testConn))
	t.Cleanup(func() {
		err := page.Context().ClearCookies()
		if err != nil {
			t.Fatalf("could not clear cookies: %v", err)
		}
	})
	return ctx, testConn
}
