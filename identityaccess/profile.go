package identityaccess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jm96441n/movieswithfriends/identityaccess/store"
	"golang.org/x/crypto/bcrypt"
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
	CreatedAt time.Time
	Stats     ProfileStats
	Account   Account
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

func (p *ProfileService) GetProfileByID(ctx context.Context, profileID int) (*Profile, error) {
	profile, err := p.db.GetProfileByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			return &Profile{}, err
		}
		return &Profile{}, err
	}
	prof := convertGetProfileResultToProfile(profile)
	prof.db = p.db

	return prof, nil
}

func (p *ProfileService) CreateProfile(ctx context.Context, logger *slog.Logger, req SignupReq) (Profile, error) {
	err := req.Validate()
	if err != nil {
		return Profile{}, err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		logger.Error("error hashing password", slog.Any("error", err))
		return Profile{}, err
	}

	res, err := p.db.CreateProfile(ctx, req.Email, req.FirstName, req.LastName, hashedPassword)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateEmailAddress) {
			logger.Debug("email exists for account")
			return Profile{}, ErrAccountExists
		}
		logger.Error("error creating account", slog.Any("error", err))
		return Profile{}, err
	}
	return Profile{
		ID:        res.ProfileID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Account: Account{
			ID:    res.AccountID,
			Email: req.Email,
		},
	}, nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

	err = p.db.UpdateProfile(ctx, updateAccountAttrs, updateProfileAttrs)
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

type ProfileParty struct {
	ID          int
	Name        string
	MemberCount int
	MovieCount  int
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

func convertGetProfileResultToProfile(res store.GetProfileResult) *Profile {
	return &Profile{
		ID:        res.ID,
		FirstName: res.FirstName,
		LastName:  res.LastName,
		CreatedAt: res.CreatedAt,
		Account: Account{
			ID:       res.AccountID,
			Email:    res.AccountEmail,
			Password: res.AccountPassword,
		},
	}
}
