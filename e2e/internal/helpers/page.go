package helpers

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

func FillInField(t *testing.T, label, value string, page playwright.Page) {
	t.Helper()
	field := page.GetByLabel(label)

	if field == nil {
		t.Fatalf("could not find field labeled by %q", label)
	}

	err := field.Fill(value)
	if err != nil {
		t.Fatalf("could not fill %s: %v", label, err)
	}
}

func OpenPage(t *testing.T, browser playwright.Browser, addr string) playwright.Page {
	t.Helper()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto(addr); err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	return page
}
