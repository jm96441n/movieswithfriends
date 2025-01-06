package partymgmt

import (
	"context"

	"github.com/jm96441n/movieswithfriends/store"
)

type MemberService struct {
	db *store.PGStore
}

func NewMemberService(db *store.PGStore) *MemberService {
	return &MemberService{db: db}
}

func (s *MemberService) GetWatchHistory(ctx context.Context, memberID, offset int) ([]store.WatchedMoviesForMemberResult, int, error) {
	watchedMovies, err := s.db.GetWatchedMoviesForMember(ctx, memberID, offset)
	if err != nil {
		return nil, 0, err
	}

	numRecords, err := s.db.GetWatchedMoviesCountForMember(ctx, memberID)
	if err != nil {
		return nil, 0, err
	}

	return watchedMovies, numRecords, nil
}
