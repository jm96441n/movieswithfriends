package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func TestMovieSearch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(*testing.T){
		"testSearchSuccessfulSearchFromSearchPage": testSearchSuccessfulSearchFromSearchPage(ctx, connPool, page, port),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testSearchSuccessfulSearchFromSearchPage(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s", appPort))
		helpers.Ok(t, err, "could not go to the index page")

		err = page.Locator("#nav-search").Click()
		helpers.Ok(t, err, "could not click search button")

		curURL := page.URL()
		helpers.Assert(t, strings.Contains(curURL, "/movies"), "expected to be on movie search page, got %s", curURL)

		helpers.FillInField(t, helpers.FormField{Label: "", Value: "The Matrix"}, page)
		page.Keyboard().Press("Enter")
		asserter := playwright.NewPlaywrightAssertions()

		matrixOne := page.Locator("#the-matrix").First()
		matrixTwo := page.Locator("#the-matrix-reloaded").First()
		matrixThree := page.Locator("#the-matrix-revolutions").First()

		helpers.Ok(t, asserter.Locator(matrixOne).ToBeVisible(), "Matrix One Card is not visible")
		helpers.Ok(t, asserter.Locator(matrixTwo).ToBeVisible(), "Matrix Reloaded Card is not visible")
		helpers.Ok(t, asserter.Locator(matrixThree).ToBeVisible(), "Matrix Revolutions Card is not visible")

		cases := map[string]playwright.Locator{
			"The Matrix":             matrixOne,
			"The Matrix Reloaded":    matrixTwo,
			"The Matrix Revolutions": matrixThree,
		}
		for title, locator := range cases {
			err = locator.Locator("text='Details'").Click()
			helpers.Ok(t, err, "could not click Details button for '%s'", title)

			curURL = page.URL()
			helpers.Assert(t, strings.Contains(curURL, "/movies/"), "expected to be on movie detail page, got %s", curURL)

			helpers.Ok(t, asserter.Locator(page.Locator("#title")).ToHaveText(title), "'%s' title is not on page", title)

			page.GoBack()

			helpers.Ok(t, asserter.Locator(matrixOne).ToBeVisible(), "Matrix One Card is not visible")
			helpers.Ok(t, asserter.Locator(matrixTwo).ToBeVisible(), "Matrix Reloaded Card is not visible")
			helpers.Ok(t, asserter.Locator(matrixThree).ToBeVisible(), "Matrix Revolutions Card is not visible")

			curURL = page.URL()
			helpers.Assert(t, strings.Contains(curURL, "/movies"), "expected to be on movie search page, got %s", curURL)
		}
	}
}
