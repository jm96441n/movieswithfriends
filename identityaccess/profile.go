package identityaccess

import (
	"context"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
)

type Profile struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	Stats     ProfileStats
}

type ProfileStats struct {
	NumberOfParties int
	WatchTime       int
	MoviesWatched   int
}

type ProfileRepository interface {
	GetProfileByID(context.Context, int) (store.Profile, error)
	GetProfileStats(context.Context, int) (int, int, int, error)
	GetAccountEmailForProfile(context.Context, int) (string, error)
}

type ProfileService struct {
	db ProfileRepository
}

func NewProfileService(db ProfileRepository) *ProfileService {
	return &ProfileService{db: db}
}

func (p *ProfileService) GetProfileByID(ctx context.Context, id int) (Profile, error) {
	profile, err := p.db.GetProfileByID(ctx, id)
	if err != nil {
		return Profile{}, err
	}

	numParties, watchTime, moviesWatched, err := p.db.GetProfileStats(ctx, id)
	if err != nil {
		return Profile{}, err
	}

	email, err := p.db.GetAccountEmailForProfile(ctx, id)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		ID:        profile.ID,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Email:     email,
		CreatedAt: profile.CreatedAt,
		Stats: ProfileStats{
			NumberOfParties: numParties,
			WatchTime:       watchTime,
			MoviesWatched:   moviesWatched,
		},
	}, nil
}
