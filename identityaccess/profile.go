package identityaccess

import (
	"context"
	"time"

	"github.com/jm96441n/movieswithfriends/store"
)

type Account struct {
	Email    string
	Password []byte
}

type Profile struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	Stats     ProfileStats
	AccountID int
}

type ProfileUpdateReq struct {
	FirstName               string
	LastName                string
	Email                   string
	CurrentPassword         string
	NewPassword             string
	NewPasswordConfirmation string
}

type ProfileStats struct {
	NumberOfParties int
	WatchTime       int
	MoviesWatched   int
}

type ProfileService struct {
	db *store.PGStore
}

func NewProfileService(db *store.PGStore) *ProfileService {
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

func (p *ProfileService) UpdateProfile(ctx context.Context, req ProfileUpdateReq, profile Profile) error {
	return nil
}
