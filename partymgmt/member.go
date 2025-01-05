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

func (s *MemberService) GetWatchHistory(ctx context.Context, memberID, offset int) ([]store.WatchedMoviesForMemberResult, error) {
	return s.db.GetWatchedMoviesForMember(ctx, memberID, offset)
}
