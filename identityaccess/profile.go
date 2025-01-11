package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess/store"
)

type Account struct {
	ID       int
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
	Account   Account
	AccountID int
	db        *store.ProfileRepository
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
	db *store.ProfileRepository
}

func NewProfileService(db *store.ProfileRepository) *ProfileService {
	return &ProfileService{db: db}
}

func (p *ProfileService) GetProfileByIDWithStats(ctx context.Context, logger *slog.Logger, id int) (*Profile, error) {
	profile, err := p.db.GetProfileByID(ctx, id)
	if err != nil {
		return nil, err
	}

	numParties, watchTime, moviesWatched, err := p.db.GetProfileStats(ctx, logger, id)
	if err != nil {
		return nil, err
	}

	return &Profile{
		ID:        profile.ID,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		CreatedAt: profile.CreatedAt,
		db:        p.db,
		Stats: ProfileStats{
			NumberOfParties: numParties,
			WatchTime:       watchTime,
			MoviesWatched:   moviesWatched,
		},
	}, nil
}

func (p *ProfileService) GetProfileByID(ctx context.Context, profileID int) (Profile, error) {
	profile, err := p.db.GetProfileByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			return Profile{}, err
		}
		return Profile{}, err
	}

	return convertResult(profile), nil
}

func convertResult(profile store.GetProfileByIDResult) Profile {
	return Profile{
		ID:        profile.ID,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		CreatedAt: profile.CreatedAt,
		Account: Account{
			ID:    profile.AccountID,
			Email: profile.AccountEmail,
		},
	}
}

func (p *Profile) Update(ctx context.Context, logger *slog.Logger, req ProfileUpdateReq) error {
	err := validateUpdateRequest(req)
	if err != nil {
		return err
	}

	updateProfileAttrs := store.ProfileUpdateAttrs{
		ID:        p.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	updateAccountAttrs := store.AccountUpdateAttrs{
		ID:    p.Account.ID,
		Email: req.Email,
	}

	if req.NewPassword != "" {
		pw, err := hashPassword(req.NewPassword)
		if err != nil {
			logger.Error("error hashing password", slog.Any("error", err))
			return err
		}
		logger.Debug("setting new password")
		updateAccountAttrs.Password = pw
	}

	err = p.db.UpdateProfileAndAccount(ctx, updateAccountAttrs, updateProfileAttrs)
	if err != nil {
		logger.Error("error updating profile and account", slog.Any("error", err))
		return err
	}

	p.FirstName = req.FirstName
	p.LastName = req.LastName
	p.Account.Email = req.Email

	logger.Info("updated profile")

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
