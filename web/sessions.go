package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jm96441n/movieswithfriends/metrics"
)

// Session helper methods

const sessionName = "moviesWithFriendsCookie"

type SessionAccessor struct {
	sessionStore *sessions.CookieStore
	logger       *slog.Logger
}

type flashes struct {
	errorFlashes   []interface{}
	warningFlashes []interface{}
	infoFlashes    []interface{}
}

var (
	ErrFailedToGetProfileIDFromSession = errors.New("failed to get profile id from session")
	ErrFailedToGetPartyIDFromSession   = errors.New("failed to get current party id from session")
	ErrFailedToGetAccountIDFromSession = errors.New("failed to get accountl id from session")
)

func NewSessionAccessor(sessionStore *sessions.CookieStore, logger *slog.Logger) *SessionAccessor {
	return &SessionAccessor{
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (s *SessionAccessor) getFlashes(ctx context.Context, r *http.Request, w http.ResponseWriter) flashes {
	session, err := s.sessionStore.Get(r, sessionName)
	var flashes flashes
	if err == nil {
		flashes.errorFlashes = session.Flashes(FlashErrorKey)
		flashes.warningFlashes = session.Flashes(FlashWarningKey)
		flashes.infoFlashes = session.Flashes(FlashInfoKey)
		session.Save(r, w)
	}
	return flashes
}

// getProfileIDFromSession retrieves the profile ID from the session.
func (s *SessionAccessor) GetProfileIDFromSession(ctx context.Context, r *http.Request) (int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Handlers.getProfileIDFromSession")
	defer span.End()

	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		session, err = s.sessionStore.New(r, sessionName)
		if err != nil {
			s.logger.DebugContext(ctx, "failed to create new session")
			return 0, nil
		}
	}

	sessionProfileID := session.Values["profileID"]
	profileID, ok := sessionProfileID.(int)
	if !ok {
		return 0, ErrFailedToGetProfileIDFromSession
	}

	return profileID, nil
}

func (s *SessionAccessor) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx, span, _ := metrics.SpanFromContext(r.Context(), "Handlers.logout")
	defer span.End()

	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get session", slog.Any("error", err))
		return err
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to save session", slog.Any("error", err))
		return err
	}
	return nil
}

func (s *SessionAccessor) GetAccountIDFromSession(ctx context.Context, r *http.Request) (int, error) {
	_, span, _ := metrics.SpanFromContext(ctx, "getAccountIDFromSession")
	defer span.End()

	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		session, err = s.sessionStore.New(r, sessionName)
		if err != nil {
			s.logger.DebugContext(ctx, "failed to create new session")
			return 0, nil
		}
	}

	sessionAccountID := session.Values["accountID"]
	accountID, ok := sessionAccountID.(int)
	if !ok {
		return 0, ErrFailedToGetAccountIDFromSession
	}

	return accountID, nil
}

func (s *SessionAccessor) SetProfileSessionValues(ctx context.Context, w http.ResponseWriter, r *http.Request, accountID int, profileID int, fullName string, email string) error {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		s.logger.ErrorContext(ctx, "error getting session from store success path", slog.Any("error", err))
		return err
	}

	session.Values["accountID"] = accountID
	session.Values["profileID"] = profileID
	session.Values["fullName"] = fullName
	session.Values["email"] = email

	err = session.Save(r, w)
	if err != nil {
		s.logger.ErrorContext(ctx, "error saving session", slog.Any("error", err))
		return err
	}
	return nil
}
