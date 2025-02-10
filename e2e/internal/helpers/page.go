package helpers

import (
	"errors"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/playwright-community/playwright-go"
)

func WaitForInput(t *testing.T) {
	t.Helper()
	listener, err := net.Listen("tcp", ":0")
	Ok(t, err, "could not start listener")

	mux := http.NewServeMux()
	port := listener.Addr().(*net.TCPAddr).Port
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("input received\n"))
		listener.Close()
		t.Log("input received, continuing test")
	})

	t.Logf("Waiting for input on http://localhost:%d", port)
	if err := http.Serve(listener, mux); !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}

func LoginAs(t *testing.T, page playwright.Page, info TestAccountInfo) {
	t.Helper()
	sessionKey := os.Getenv("SESSION_KEY")
	codecs := securecookie.CodecsFromPairs([]byte(sessionKey))
	value, err := securecookie.EncodeMulti("moviesWithFriendsCookie", map[interface{}]interface{}{
		"accountID": info.AccountID,
		"profileID": info.ProfileID,
		"fullName":  info.FirstName + " " + info.LastName,
		"email":     info.Email,
	}, codecs...)

	Ok(t, err, "could not encode cookie")

	page.Context().AddCookies([]playwright.OptionalCookie{
		{
			Name:     "moviesWithFriendsCookie",
			Domain:   playwright.String("localhost"),
			Value:    value,
			Path:     playwright.String("/"),
			SameSite: playwright.SameSiteAttributeNone,
			Secure:   playwright.Bool(true),
		},
	})
}

type FormField struct {
	Label string
	Value string
}

func FillInField(t *testing.T, ff FormField, page playwright.Page, options ...playwright.PageGetByLabelOptions) {
	t.Helper()
	field := page.GetByLabel(ff.Label, options...)

	Assert(t, field != nil, "could not find field labeled by %q", ff.Label)

	err := field.Fill(ff.Value)
	Ok(t, err, "could not fill %s", ff.Label)
}

func LocatorHasText(t *testing.T, page playwright.Page, pageAssertions playwright.PlaywrightAssertions, locator string, text string) {
	node := page.Locator(locator)
	Assert(t, node != nil, "could not get element at: %q", locator)

	err := pageAssertions.Locator(node).ToHaveText(text)
	Ok(t, err, "expected node %q to have text %q, got %v", locator, text, err)
}

func InfoFlashMessageShouldBe(t *testing.T, page playwright.Page, pageAssertions playwright.PlaywrightAssertions, message string) {
	t.Helper()
	locator := page.Locator(".alert-primary")

	Assert(t, locator != nil, "could not find info flash message")

	Ok(t, pageAssertions.Locator(locator).ToHaveText(message), "expected info flash message to be %q", message)
}
