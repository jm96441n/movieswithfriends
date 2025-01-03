package partymgmt

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"

	"github.com/jm96441n/movieswithfriends/store"
)

type partyStore interface {
	CreateParty(context.Context, int, string, string) (int, error)
	GetPartyByShortID(context.Context, string) (store.Party, error)
	CreatePartyMember(context.Context, int, int) error
	GetPartyByIDWithStats(context.Context, int) (store.GetPartyByIDWithStatsResult, error)
}

type movieStore interface {
	GetMoviesForParty(context.Context, int, int) (store.MoviesByStatus, error)
}

// TOOD: should this exist? maybe just pass db to functions that need it
type PartyService struct {
	Logger           *slog.Logger
	DB               partyStore
	MoviesRepository movieStore
}

type Party struct {
	ID              int
	Name            string
	ShortID         string
	MemberCount     int
	MovieCount      int
	WatchedCount    int
	WatchedMovies   []*store.WatchedMovie
	UnwatchedMovies []*store.UnwatchedMovie
	SelectedMovie   *store.SelectedMovie
}

var ErrMemberExistsInParty = errors.New("member already exists in party")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (s *PartyService) AddFriendToParty(ctx context.Context, idMember int, shortID string) error {
	party, err := s.DB.GetPartyByShortID(ctx, shortID)
	if err != nil {
		return err
	}

	err = s.DB.CreatePartyMember(ctx, idMember, party.ID)
	if errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
		return errors.Join(ErrMemberExistsInParty, err)
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *PartyService) CreateParty(ctx context.Context, idMember int, name string) (int, error) {
	successFullyCreated := false
	var (
		id  int
		err error
	)
	for i := 0; i < 5; i++ {
		shortID := generateRandomString()
		id, err = s.DB.CreateParty(ctx, idMember, name, shortID)
		if errors.Is(err, store.ErrDuplicatePartyShortID) {
			continue
		}
		if err != nil {
			return 0, err
		}

		successFullyCreated = true
		break
	}

	if !successFullyCreated {
		return 0, errors.New("failed to create party")
	}

	return id, nil
}

type partyGetter interface {
	GetPartyByIDWithStats(context.Context, int) (store.GetPartyByIDWithStatsResult, error)
}

type movieGetter interface {
	GetMoviesForParty(context.Context, int, int) ([]store.Movie, error)
}

func (s *PartyService) GetPartyWithMovies(ctx context.Context, id int) (Party, error) {
	results, err := s.DB.GetPartyByIDWithStats(ctx, id)
	if err != nil {
		s.Logger.Error("failed to get party by id", slog.Any("error", err))
		return Party{}, err
	}

	party := Party{
		ID:           results.ID,
		Name:         results.Name,
		ShortID:      results.ShortID,
		MemberCount:  results.MemberCount,
		MovieCount:   results.MovieCount,
		WatchedCount: results.WatchedCount,
	}

	moviesByStatus, err := s.MoviesRepository.GetMoviesForParty(ctx, party.ID, 0)
	if err != nil {
		s.Logger.Error("failed to get movies for party", slog.Any("error", err))
		return Party{}, err
	}

	party.WatchedMovies = moviesByStatus.WatchedMovies
	party.UnwatchedMovies = moviesByStatus.UnwatchedMovies
	party.SelectedMovie = moviesByStatus.SelectedMovie

	return party, nil
}

// generate a random 6 character string
func generateRandomString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
