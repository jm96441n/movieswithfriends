package partymgmt

import (
	"context"

	"github.com/jm96441n/movieswithfriends/store"
)

type memberStore interface {
	GetWatchedMoviesForMember(context.Context, int, int) ([]store.WatchedMoviesForMemberResult, error)
}

type MemberService struct {
	db memberStore
}

func NewMemberService(db memberStore) *MemberService {
	return &MemberService{db: db}
}

func (s *MemberService) GetWatchHistory(ctx context.Context, memberID, offset int) ([]store.WatchedMoviesForMemberResult, error) {
	return s.db.GetWatchedMoviesForMember(ctx, memberID, offset)
}
