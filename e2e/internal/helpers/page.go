package helpers

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

func FillInField(t *testing.T, label, value string, page playwright.Page) {
	t.Helper()
	field := page.GetByLabel(label)

	Assert(t, field != nil, "could not find field labeled by %q", label)

	err := field.Fill(value)
	Ok(t, err, "could not fill %s", label)
}
