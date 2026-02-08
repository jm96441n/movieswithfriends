package e2e_test

import (
	"context"
	"testing"

	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
)

func TestMovieShow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testConn, page, _ := helpers.SetupSuite(ctx, t)

	helpers.Setup(ctx, t, testConn, page)
}
