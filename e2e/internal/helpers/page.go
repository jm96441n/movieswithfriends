package helpers

import (
	"os"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/playwright-community/playwright-go"
)

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

func FillInField(t *testing.T, ff FormField, page playwright.Page) {
	t.Helper()
	field := page.GetByLabel(ff.Label)

	Assert(t, field != nil, "could not find field labeled by %q", ff.Label)

	err := field.Fill(ff.Value)
	Ok(t, err, "could not fill %s", ff.Label)
}
