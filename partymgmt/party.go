package partymgmt

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"

	"github.com/jm96441n/movieswithfriends/partymgmt/store"
	legacystore "github.com/jm96441n/movieswithfriends/store"
)

var ErrMemberExistsInParty = errors.New("member already exists in party")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PartyService struct {
	Logger           *slog.Logger
	DB               *store.PartyRepository
	MoviesRepository *legacystore.PGStore
}

type MovieSet map[*Movie]struct{}

type Party struct {
	ID              int
	Name            string
	ShortID         string
	MemberCount     int
	MovieCount      int
	WatchedCount    int
	WatchedMovies   MovieSet
	UnwatchedMovies MovieSet
	SelectedMovie   MovieSet
	db              *store.PartyRepository
}

func (p *Party) AddMember(ctx context.Context, idMember int) error {
	err := p.db.CreatePartyMember(ctx, idMember, p.ID)

	if errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
		return errors.Join(ErrMemberExistsInParty, err)
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *PartyService) AddNewMemberToParty(ctx context.Context, idMember int, shortID string) error {
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

	// TODO: FINISH THIS REFACTOR
	party.WatchedMovies = moviesByStatus.WatchedMovies
	party.UnwatchedMovies = moviesByStatus.UnwatchedMovies
	party.SelectedMovie = moviesByStatus.SelectedMovie

	return party, nil
}

func (p Party) GetMoviesByStatus(ctx context.Context) (MovieSet, MovieSet, MovieSet, error) {
	moviesByStatus, err := p.db.GetMoviesByStatus(ctx, p.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	return moviesByStatus.WatchedMovies, moviesByStatus.UnwatchedMovies, moviesByStatus.SelectedMovie, nil
}

func (s *PartyService) GetPartyByShortID(ctx context.Context, shortID string) (Party, error) {
	res, err := s.DB.GetPartyByShortID(ctx, shortID)
	if err != nil {
		return Party{}, err
	}

	return Party{
		ID:      res.ID,
		Name:    res.Name,
		ShortID: res.ShortID,
	}, nil
}

// generate a random 6 character string
func generateRandomString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
