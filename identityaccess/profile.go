package identityaccess

import (
	"context"
	"errors"
	"fmt"
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
		AccountID: profile.AccountID,
		Stats: ProfileStats{
			NumberOfParties: numParties,
			WatchTime:       watchTime,
			MoviesWatched:   moviesWatched,
		},
	}, nil
}

func (p *ProfileService) UpdateProfile(ctx context.Context, req ProfileUpdateReq, profile Profile) error {
	err := validateUpdateRequest(req)
	if err != nil {
		return err
	}

	if req.NewPassword == "" {
		err = p.db.UpdateProfile(ctx, req.FirstName, req.LastName, req.Email, profile.ID)
		if err != nil {
			return err
		}

		err = p.db.UpdateAccountEmail(ctx, req.Email, profile.AccountID)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

var (
	ErrFirstNameIsRequired            = errors.New("first name is required")
	ErrLastNameIsRequired             = errors.New("last name is required")
	ErrEmailIsRequired                = errors.New("email is required")
	ErrNewPasswordMustMatchIsRequired = errors.New("email is required")
)

type ProfileEditValidationError struct {
	EmailError            error
	PasswordError         error
	NewPasswordMatchError error
	FirstNameError        error
	LastNameError         error
}

func (s *ProfileEditValidationError) Error() string {
	return fmt.Sprintf("profile edit validation error: %#v", s)
}

func (s *ProfileEditValidationError) IsNil() bool {
	return s.EmailError == nil && s.PasswordError == nil && s.NewPasswordMatchError == nil && s.FirstNameError == nil && s.LastNameError == nil
}

func validateUpdateRequest(req ProfileUpdateReq) error {
	var err ProfileEditValidationError
	if req.FirstName == "" {
		err.FirstNameError = ErrFirstNameIsRequired
	}

	if req.LastName == "" {
		err.LastNameError = ErrLastNameIsRequired
	}

	if req.Email == "" {
		err.EmailError = ErrEmailIsRequired
	}

	if !err.IsNil() {
		return &err
	}

	return nil
}
