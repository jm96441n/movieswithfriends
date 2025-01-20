package partymgmt

import (
	"context"
	"log/slog"

	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type WatcherService struct {
	db *store.WatcherRepository
}

type Watcher struct {
	ID int
	db *store.WatcherRepository
}

func NewWatcherService(db *store.WatcherRepository) *WatcherService {
	return &WatcherService{db: db}
}

func (s *WatcherService) GetWatcher(ctx context.Context, memberID int) (Watcher, error) {
	return Watcher{
		ID: memberID,
		db: s.db,
	}, nil
}

func (s *WatcherService) GetWatchHistory(ctx context.Context, logger *slog.Logger, memberID, offset int) ([]store.WatchedMoviesForWatcherResult, int, error) {
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

func (w *Watcher) GetWatchHistory(ctx context.Context, logger *slog.Logger, memberID, offset int) ([]store.WatchedMoviesForWatcherResult, int, error) {
	watchedMovies, err := w.db.GetWatchedMoviesForWatcher(ctx, memberID, offset)
	if err != nil {
		return nil, 0, err
	}

	numRecords, err := w.db.GetWatchedMoviesCountForMember(ctx, logger, w.ID)
	if err != nil {
		return nil, 0, err
	}

	return watchedMovies, numRecords, nil
}

func (w *Watcher) GetParties(ctx context.Context) ([]Party, error) {
	parties, err := w.db.GetPartiesForWatcher(ctx, w.ID, 50)
	if err != nil {
		return nil, err
	}

	var res []Party
	for _, party := range parties {
		res = append(res, Party{
			ID:          party.ID,
			Name:        party.Name,
			MemberCount: party.MemberCount,
			MovieCount:  party.MovieCount,
		})
	}

	return res, nil
}

func (w *Watcher) GetCurrentPartyID(ctx context.Context) (int, error) {
	parties, err := w.db.GetPartiesForWatcher(ctx, w.ID, 1)
	if err != nil {
		return 0, err
	}

	if len(parties) == 0 {
		return 0, nil
	}

	return parties[0].ID, nil
}

type PartiesForMovie struct {
	WithMovie    []Party
	WithoutMovie []Party
}

func (w *Watcher) GetPartiesToAddMovie(ctx context.Context, logger *slog.Logger, idMovie int) (PartiesForMovie, error) {
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

	err = w.db.GetWatcherPartiesWithoutMovie(ctx, logger, w.ID, idMovie, func(partyID int, partyName string) {
		p := Party{
			Name: partyName,
			ID:   partyID,
		}
		parties.WithoutMovie = append(parties.WithoutMovie, p)
	})
	if err != nil {
		return PartiesForMovie{}, err
	}
	return parties, nil
}
