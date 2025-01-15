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

func (p *Watcher) GetWatchHistory(ctx context.Context, logger *slog.Logger, memberID, offset int) ([]store.WatchedMoviesForWatcherResult, int, error) {
	watchedMovies, err := p.db.GetWatchedMoviesForWatcher(ctx, memberID, offset)
	if err != nil {
		return nil, 0, err
	}

	numRecords, err := p.db.GetWatchedMoviesCountForMember(ctx, logger, p.ID)
	if err != nil {
		return nil, 0, err
	}

	return watchedMovies, numRecords, nil
}

func (p *Watcher) GetParties(ctx context.Context) ([]Party, error) {
	parties, err := p.db.GetPartiesForWatcher(ctx, p.ID)
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
