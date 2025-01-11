package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/e2e/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

func TestProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	connPool, page, port := helpers.SetupSuite(ctx, t)

	tests := map[string]func(*testing.T){
		"testCanViewProfile": testCanViewProfile(ctx, connPool, page, port),
		"testCanEditProfile": testCanEditProfile(ctx, connPool, page, port),
	}

	for name, testFn := range tests {
		t.Run(name, testFn)
	}
}

func testCanViewProfile(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		currentAccount := helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "buddy@santa.com", Password: "anotherpassword", FirstName: "Buddy", LastName: "TheElf"})
		setupProfileViewData(ctx, t, testConn, currentAccount)
		helpers.LoginAs(t, page, currentAccount)

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/profile", appPort))
		helpers.Ok(t, err, "could not goto profile page")

		pageAssertions := playwright.NewPlaywrightAssertions()

		// Check the counts for movies watched
		helpers.LocatorHasText(t, page, pageAssertions, "#count-watched-movies", "16") // 7 movies from party1, 6 movies from party2, 3 movies from party3 -> 16 movies watched

		// Check the counts for joined parties
		helpers.LocatorHasText(t, page, pageAssertions, "#count-joined-parties", "3") // 3 parties joined

		// Check the watch time
		helpers.LocatorHasText(t, page, pageAssertions, "#watch-time", "31h 17m") // (7 movies * 125 minutes each) + (6 movies * 120 minutes each) + (3 movies * 94 minutes each) = 875min + 720min + 282min = 1877min = 31h 17m

		// Check the number of parties being shown
		parties := page.Locator(".party-card")
		helpers.Assert(t, parties != nil, "could not find items in .party-card")

		err = pageAssertions.Locator(parties).ToHaveCount(4) // 3 parties and 1 section for the "Create New Party" card
		helpers.Ok(t, err, "expected 3 parties and the 'Create Party' card, got %v", err)

		// check that we paginate watched movies to 5 in a list, we should see 5 currently and when we hit the next page button we should see 3
		recentlyWatchedMovies := page.Locator(".recently-watched-movie")
		helpers.Assert(t, recentlyWatchedMovies != nil, "could not find recently watched movies in .recently-watched-movie")

		err = pageAssertions.Locator(recentlyWatchedMovies).ToHaveCount(5)
		helpers.Ok(t, err, "expected 5 recently watched movies, got %v", err)

		// check the pagination shows correctly
		pagination := page.Locator(".pagination")
		helpers.Assert(t, pagination != nil, "could not find pagination in .pagination")

		err = pageAssertions.Locator(pagination).ToHaveText("First 1 2 3 Last")
		helpers.Ok(t, err, "expected pagination to have text 'First 1 2 3 Last', got %v", err)

		// go to third page of pagination and check that there are 5 videos in the list and the pagination shows correctly
		helpers.Ok(t, pagination.Locator("text=3").Click(), "could not click button 3 in pagination")

		err = pageAssertions.Locator(pagination).ToHaveText("First 2 3 4 Last")
		helpers.Ok(t, err, "expected pagination to have text 'First 2 3 4 Last', got %v", err)

		err = pageAssertions.Locator(recentlyWatchedMovies).ToHaveCount(5)
		helpers.Ok(t, err, "expected 5 recently watched movies, got %v", err)

		// click the Last button and show that pagination shows correctly
		helpers.Ok(t, pagination.Locator("text=Last").Click(), "could not click 'Last' button in pagination")

		err = pageAssertions.Locator(pagination).ToHaveText("First 2 3 4 Last")
		helpers.Ok(t, err, "expected pagination to have text 'First 2 3 4 Last', got %v", err)

		err = pageAssertions.Locator(recentlyWatchedMovies).ToHaveCount(1)
		helpers.Ok(t, err, "expected 1 recently watched movies, got %v", err)

		// click the First button and show that pagination shows correctly
		helpers.Ok(t, pagination.Locator("text=First").Click(), "could not click 'First' button in pagination")

		err = pageAssertions.Locator(pagination).ToHaveText("First 1 2 3 Last")
		helpers.Ok(t, err, "expected pagination to have text 'First 1 2 3 Last', got %v", err)

		err = pageAssertions.Locator(recentlyWatchedMovies).ToHaveCount(5)
		helpers.Ok(t, err, "expected 5 recently watched movies, got %v", err)
	}
}

func testCanEditProfile(ctx context.Context, testConn *pgxpool.Pool, page playwright.Page, appPort string) func(t *testing.T) {
	return func(t *testing.T) {
		helpers.Setup(ctx, t, testConn, page)
		currentAccount := helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{
			Email:     "buddy@santa.com",
			Password:  "anotherpassword",
			FirstName: "Buddy",
			LastName:  "TheElf",
		})

		helpers.LoginAs(t, page, currentAccount)

		pageAssertions := playwright.NewPlaywrightAssertions()

		_, err := page.Goto(fmt.Sprintf("http://localhost:%s/profile", appPort))
		helpers.Ok(t, err, "could not goto profile page")

		newName := "NewName"
		newLastName := "NewLastName"

		nameCases := []struct {
			expectedNameValue string
			formField         helpers.FormField
		}{
			{
				expectedNameValue: fmt.Sprintf("%s TheElf", newName),
				formField:         helpers.FormField{Label: "First Name", Value: newName},
			},
			{
				expectedNameValue: fmt.Sprintf("%s %s", newName, newLastName),
				formField:         helpers.FormField{Label: "Last Name", Value: newLastName},
			},
		}

		// Test the profile fields for changes
		for _, field := range nameCases {
			// navigate to profile edit page
			helpers.Ok(t, page.Locator("text=Edit Profile").Click(), "failed to click the 'Edit Profile' link")

			curURL := page.URL()
			helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile/edit", appPort), "expected to be on the edit profile page, got %s", curURL)

			// fill in specific field
			helpers.FillInField(t, field.formField, page)

			// persist changes
			helpers.Ok(t, page.Locator("button:has-text('Save Changes')").Click(), "Could not click the 'Save Changes' button")

			curURL = page.URL()
			helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile", appPort), "expected to be on the profile page, got %s", curURL)

			helpers.Ok(t, pageAssertions.Locator(page.Locator("#profile-name")).ToHaveText(field.expectedNameValue), "expected profile name field to have value %s", field.expectedNameValue)

			helpers.InfoFlashMessageShouldBe(t, page, pageAssertions, "Edited your profile!")
		}

		newEmail := "new@email.com"
		newPassword := "1NewPassword"

		// check email update

		helpers.Ok(t, page.Locator("text=Edit Profile").Click(), "failed to click the 'Edit Profile' link")

		curURL := page.URL()
		helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile/edit", appPort), "expected to be on the edit profile page, got %s", curURL)

		// fill in email field
		helpers.FillInField(t, helpers.FormField{Value: newEmail, Label: "Email Address"}, page)

		// persist changes
		helpers.Ok(t, page.Locator("button:has-text('Save Changes')").Click(), "Could not click the 'Save Changes' button")

		curURL = page.URL()
		helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile", appPort), "expected to be on the profile page, got %s", curURL)

		// time.Sleep(30 * time.Minute)
		// logout and then login to try the new email
		logoutViaDropdown(t, page)

		// login with the new email
		loginThroughUI(t, page, newEmail, "anotherpassword")

		// check password update

		helpers.Ok(t, page.Locator("text=Edit Profile").Click(), "failed to click the 'Edit Profile' link")

		curURL = page.URL()
		helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile/edit", appPort), "expected to be on the edit profile page, got %s", curURL)

		// fill in password fields
		helpers.FillInField(t, helpers.FormField{Value: "anotherpassword", Label: "Current Password"}, page)
		helpers.FillInField(t, helpers.FormField{Value: newPassword, Label: "New Password"}, page, playwright.PageGetByLabelOptions{Exact: playwright.Bool(true)})
		helpers.FillInField(t, helpers.FormField{Value: newPassword, Label: "Confirm New Password"}, page)

		// persist changes
		helpers.Ok(t, page.Locator("button:has-text('Save Changes')").Click(), "Could not click the 'Save Changes' button")

		curURL = page.URL()
		helpers.Assert(t, curURL == fmt.Sprintf("http://localhost:%s/profile", appPort), "expected to be on the profile page, got %s", curURL)

		// logout and then login to try the new email
		logoutViaDropdown(t, page)

		// login with the new email
		loginThroughUI(t, page, newEmail, newPassword)
	}
}

func logoutViaDropdown(t *testing.T, page playwright.Page) {
	t.Helper()
	err := page.Locator(".dropdown > #user-nav-dropdown-btn").Click()
	helpers.Ok(t, err, "failed to click dropown button")

	err = page.Locator("text=Sign Out").Click()
	helpers.Ok(t, err, "failed to click link to sign out")
}

func loginThroughUI(t *testing.T, page playwright.Page, email, password string) {
	t.Helper()

	curURL := page.URL()
	helpers.Assert(t, strings.Contains(curURL, "/login"), "expected to be on login page, got %s", curURL)

	helpers.FillInField(t, helpers.FormField{Label: "Email Address", Value: email}, page)
	helpers.FillInField(t, helpers.FormField{Label: "Password", Value: password}, page)

	err := page.Locator("button:has-text('Sign In')").Click()
	helpers.Ok(t, err, "could not click Sign In button")

	curURL = page.URL()
	helpers.Assert(t, strings.Contains(curURL, "/profile"), "expected to be on profile page, got %s", curURL)
}

func setupProfileViewData(ctx context.Context, t *testing.T, testConn *pgxpool.Pool, currentAccount helpers.TestAccountInfo) {
	// user not in any party
	helpers.SeedAccountWithProfile(ctx, t, testConn, helpers.TestAccountInfo{Email: "anotherUser@gmail.com", Password: "anotherpassword", FirstName: "Another", LastName: "User"})
	partyCfgs := []helpers.PartyConfig{
		{
			NumMembers:       2,
			NumMovies:        8,
			NumWatchedMovies: 7,
			MovieRuntime:     125,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       3,
			NumMovies:        9,
			NumWatchedMovies: 6,
			MovieRuntime:     120,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       1,
			NumMovies:        5,
			NumWatchedMovies: 3,
			MovieRuntime:     94,
			CurrentAccount:   currentAccount,
		},
		{
			NumMembers:       2,
			NumMovies:        5,
			NumWatchedMovies: 3,
			MovieRuntime:     120,
		},
	}

	for _, partyCfg := range partyCfgs {
		helpers.SeedPartyWithUsersAndMovies(ctx, t, testConn, partyCfg)
	}
}
