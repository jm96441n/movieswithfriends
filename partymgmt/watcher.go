package partymgmt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type WatcherService struct {
	db *store.WatcherRepository
}

type Watcher struct {
	ID    int
	Email string
	db    *store.WatcherRepository
}

var ErrWatcherNotFound = errors.New("watcher not found")

func NewWatcherService(db *store.WatcherRepository) WatcherService {
	return WatcherService{db: db}
}

func (s WatcherService) NewWatcher(ctx context.Context, memberID int) (Watcher, error) {
	return Watcher{
		ID: memberID,
		db: s.db,
	}, nil
}

func (s WatcherService) GetWatcherByEmail(ctx context.Context, email string) (Watcher, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherService.GetWatcherByEmail")
	defer span.End()
	w := Watcher{db: s.db, Email: email}
	err := s.db.GetWatcherByEmail(ctx, email, func(id int) {
		w.ID = id
	})
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			return Watcher{}, fmt.Errorf("%w: %s", ErrWatcherNotFound, err)
		}

		return Watcher{}, err
	}

	return w, nil
}

func (s WatcherService) GetWatchHistory(ctx context.Context, logger *slog.Logger, memberID, offset int) ([]store.WatchedMoviesForWatcherResult, int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetWatchHistory")
	defer span.End()
	watchedMovies, err := s.db.GetWatchedMoviesForWatcher(ctx, memberID, offset)
	if err != nil {
		return nil, 0, err
	}

	numRecords, err := s.db.GetWatchedMoviesCountForMember(ctx, logger, memberID)
	if err != nil {
		return nil, 0, err
	}

	return watchedMovies, numRecords, nil
}

func (w Watcher) GetWatchHistory(ctx context.Context, logger *slog.Logger, offset int) ([]store.WatchedMoviesForWatcherResult, int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetWatchHistory")
	defer span.End()
	watchedMovies, err := w.db.GetWatchedMoviesForWatcher(ctx, w.ID, offset)
	if err != nil {
		return nil, 0, err
	}

	numRecords, err := w.db.GetWatchedMoviesCountForMember(ctx, logger, w.ID)
	if err != nil {
		return nil, 0, err
	}

	return watchedMovies, numRecords, nil
}

func (w Watcher) GetPartiesAndInvitedParties(ctx context.Context, ps PartyService) ([]Party, []Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetPartiesAndInvitedParties")
	defer span.End()

	parties, err := w.GetParties(ctx, ps)
	if err != nil {
		return nil, nil, err
	}

	invitedParties, err := w.GetInvitedParties(ctx, ps)
	if err != nil {
		return nil, nil, err
	}

	return parties, invitedParties, nil
}

func (w Watcher) GetParties(ctx context.Context, ps PartyService) ([]Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetParties")
	defer span.End()

	var parties []Party
	// TODO: this should be a methon on the PartyService and take a scope object for the scope of fetching parties
	err := w.db.GetPartiesForWatcher(ctx, w.ID, 10, func(ctx context.Context, id int, name string, memberCount int, movieCount int, idOwner int) {
		parties = append(parties, ps.NewParty(ctx, id, name, movieCount, memberCount, idOwner))
	})
	if err != nil {
		return nil, err
	}

	return parties, nil
}

func (w Watcher) GetInvitedParties(ctx context.Context, ps PartyService) ([]Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetInvitedParties")
	defer span.End()

	var parties []Party
	err := w.db.GetInvitedPartiesForWatcher(ctx, w.ID, 10, func(ctx context.Context, id int, name string, memberCount int, movieCount int, idOwner int) {
		parties = append(parties, ps.NewParty(ctx, id, name, movieCount, memberCount, idOwner))
	})
	if err != nil {
		return nil, err
	}

	return parties, nil
}

type PartiesForMovie struct {
	WithMovie    []Party
	WithoutMovie []Party
}

func (w Watcher) GetPartiesToAddMovie(ctx context.Context, logger *slog.Logger, idMovie MovieID) (PartiesForMovie, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.GetPartiesToAddMovie")
	defer span.End()

	switch {
	case idMovie.TMDBID != nil:
		return w.getPartiesForMovieByTMDBID(ctx, logger, *idMovie.TMDBID)
	case idMovie.MovieID != nil:
		return w.getPartiesForMovieByMovieID(ctx, logger, *idMovie.MovieID)
	default:
		return PartiesForMovie{}, errors.New("TMDBID or MovieID must be set")
	}
}

func (w Watcher) getPartiesForMovieByTMDBID(ctx context.Context, logger *slog.Logger, tmdbID int) (PartiesForMovie, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.getPartiesForMovieByTMDBID")
	defer span.End()

	parties := PartiesForMovie{
		WithMovie:    make([]Party, 0, 10),
		WithoutMovie: make([]Party, 0, 10),
	}
	err := w.db.GetWatcherPartiesWithMovieByTMDBID(ctx, logger, w.ID, tmdbID, func(partyID int, partyName string) {
		p := Party{
			Name: partyName,
			ID:   partyID,
		}
		parties.WithMovie = append(parties.WithMovie, p)
	})
	if err != nil {
		return PartiesForMovie{}, err
	}

	err = w.db.GetWatcherPartiesWithoutMovieByTMDBID(ctx, logger, w.ID, tmdbID, func(partyID int, partyName string, movieCount int) {
		p := Party{
			Name:       partyName,
			ID:         partyID,
			MovieCount: movieCount,
		}
		parties.WithoutMovie = append(parties.WithoutMovie, p)
	})
	if err != nil {
		return PartiesForMovie{}, err
	}

	return parties, nil
}

func (w Watcher) getPartiesForMovieByMovieID(ctx context.Context, logger *slog.Logger, idMovie int) (PartiesForMovie, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Watcher.getPartiesForMovieByMovieID")
	defer span.End()

	parties := PartiesForMovie{
		WithMovie:    make([]Party, 0, 10),
		WithoutMovie: make([]Party, 0, 10),
	}
	err := w.db.GetWatcherPartiesWithMovie(ctx, logger, w.ID, idMovie, func(partyID int, partyName string) {
		p := Party{
			Name: partyName,
			ID:   partyID,
		}
		parties.WithMovie = append(parties.WithMovie, p)
	})
	if err != nil {
		return PartiesForMovie{}, err
	}

	err = w.db.GetWatcherPartiesWithoutMovie(ctx, logger, w.ID, idMovie, func(partyID int, partyName string, movieCount int) {
		p := Party{
			Name:       partyName,
			ID:         partyID,
			MovieCount: movieCount,
		}
		parties.WithoutMovie = append(parties.WithoutMovie, p)
	})
	if err != nil {
		return PartiesForMovie{}, err
	}
	return parties, nil
}
