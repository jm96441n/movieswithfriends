package web

import (
	"context"
	"net/http"

	"github.com/jm96441n/movieswithfriends/metrics"
)

const (
	FlashErrorKey   = "error"
	FlashInfoKey    = "info"
	FlashWarningKey = "warn"
)

type contextKey string

const (
	IsAuthenticatedContextKey = contextKey("isAuthenticated")
	FullNameContextKey        = contextKey("fullName")
	CurrentPartyIDContextKey  = contextKey("currentPartyID")
	EmailContextKey           = contextKey("email")
)

type BaseTemplateData struct {
	CurrentPagePath string
	ErrorFlashes    []interface{}
	InfoFlashes     []interface{}
	WarningFlashes  []interface{}
	IsAuthenticated bool
	CurrentYear     int
	FullName        string
	UserEmail       string
}

func (rr *Renderer) NewBaseTemplateData(r *http.Request, w http.ResponseWriter, path string) BaseTemplateData {
	authed := IsAuthenticated(r.Context())

	var (
		fullName string
		email    string
	)

	if authed {
		fullName = r.Context().Value(FullNameContextKey).(string)
		email = r.Context().Value(EmailContextKey).(string)
	}

	flashes := rr.sessionStore.getFlashes(r.Context(), r, w)

	return BaseTemplateData{
		ErrorFlashes:    flashes.errorFlashes,
		InfoFlashes:     flashes.infoFlashes,
		WarningFlashes:  flashes.warningFlashes,
		CurrentPagePath: path,
		CurrentYear:     2025,
		IsAuthenticated: authed,
		FullName:        fullName,
		UserEmail:       email,
	}
}

func IsAuthenticated(ctx context.Context) bool {
	ctx, span, _ := metrics.SpanFromContext(ctx, "isAuthenticated")
	defer span.End()
	isAuthenticated := ctx.Value(IsAuthenticatedContextKey)
	if isAuthenticated == nil {
		return false
	}
	return isAuthenticated.(bool)
}
