package partymgmt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jm96441n/movieswithfriends/partymgmt/store"
	"go.opentelemetry.io/otel/trace"
)

var ErrMemberExistsInParty = errors.New("member already exists in party")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PartyService struct {
	logger *slog.Logger
	db     *store.PartyRepository
}

type PartyMovie struct {
	ID          int       `json:"id_movie"`
	Title       string    `json:"title"`
	ReleaseDate string    `json:"release_date"`
	TrailerURL  string    `json:"trailer_url"`
	PosterURL   string    `json:"poster_url"`
	Runtime     int       `json:"runtime"`
	Rating      float64   `json:"vote_average"`
	Tagline     string    `json:"tagline"`
	Genres      []string  `json:"genres"`
	WatchDate   time.Time `json:"watch_date"`
	AddedBy     FullName  `json:"added_by"`
	AddedOn     time.Time `json:"created_at"`
	PartyName   string
}

type MoviesByStatus struct {
	WatchedMovies   []PartyMovie
	UnwatchedMovies []PartyMovie
	SelectedMovie   *PartyMovie
}

type PartyMember struct {
	FirstName string
	LastName  string
	ID        int
	Owner     bool
	JoinedOn  time.Time
}

type Party struct {
	ID           int
	Name         string
	ShortID      string
	Members      []PartyMember
	MemberCount  int
	MovieCount   int
	WatchedCount int

	MoviesByStatus MoviesByStatus
	db             *store.PartyRepository
}

func NewPartyService(logger *slog.Logger, db *store.PartyRepository) *PartyService {
	return &PartyService{
		logger: logger,
		db:     db,
	}
}

func (s *PartyService) NewParty(ctx context.Context) Party {
	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("PartyService").Start(ctx, "NewParty")
	defer span.End()
	return Party{db: s.db}
}

func (s *PartyService) AddNewMemberToParty(ctx context.Context, idMember int, shortID string) error {
	party, err := s.db.GetPartyByShortID(ctx, shortID)
	if err != nil {
		return err
	}

	err = s.db.CreatePartyMember(ctx, idMember, party.ID)
	if errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
		return errors.Join(ErrMemberExistsInParty, err)
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *PartyService) CreateParty(ctx context.Context, idMember int, name string) (int, error) {
	successFullyCreated := false
	var (
		id  int
		err error
	)
	for i := 0; i < 5; i++ {
		shortID := generateRandomString()
		id, err = s.db.CreateParty(ctx, idMember, name, shortID)
		if errors.Is(err, store.ErrDuplicatePartyShortID) {
			continue
		}
		if err != nil {
			return 0, err
		}

		successFullyCreated = true
		break
	}

	if !successFullyCreated {
		return 0, errors.New("failed to create party")
	}

	return id, nil
}

func (s *PartyService) GetPartyWithMovies(ctx context.Context, logger *slog.Logger, id int) (Party, error) {
	party := s.NewParty(ctx)
	err := s.db.GetPartyByIDWithStats(ctx, id, func(res store.GetPartyByIDWithStatsResult) {
		party.ID = res.ID
		party.Name = res.Name
		party.ShortID = res.ShortID
		party.MemberCount = res.MemberCount
		party.MovieCount = res.MovieCount
		party.WatchedCount = res.WatchedCount
	})
	if err != nil {
		s.logger.Error("failed to get party by id", slog.Any("error", err))
		return Party{}, err
	}

	err = party.GetPartyMembers(ctx)
	if err != nil {
		s.logger.Error("failed to get party members", slog.Any("error", err))
		return Party{}, err
	}

	moviesByStatus, err := party.GetMoviesByStatus(ctx, logger)
	if err != nil {
		s.logger.Error("failed to get movies for party", slog.Any("error", err))
		return Party{}, err
	}

	party.MoviesByStatus = moviesByStatus

	return party, nil
}

func (s *PartyService) GetPartyByShortID(ctx context.Context, shortID string) (Party, error) {
	res, err := s.db.GetPartyByShortID(ctx, shortID)
	if err != nil {
		return Party{}, err
	}

	return Party{
		ID:      res.ID,
		Name:    res.Name,
		ShortID: res.ShortID,
		db:      s.db,
	}, nil
}

func (p Party) AddMember(ctx context.Context, idMember int) error {
	err := p.db.CreatePartyMember(ctx, idMember, p.ID)

	if errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
		return errors.Join(ErrMemberExistsInParty, err)
	}

	if err != nil {
		return err
	}
	return nil
}

func (p Party) GetMoviesByStatus(ctx context.Context, logger *slog.Logger) (MoviesByStatus, error) {
	moviesByStatus := MoviesByStatus{
		WatchedMovies:   make([]PartyMovie, 0, 10),
		UnwatchedMovies: make([]PartyMovie, 0, 10),
		SelectedMovie:   nil,
	}
	err := p.db.GetMoviesForParty(ctx, logger, p.ID, 0, func(status store.WatchStatusEnum, movieJSON []byte) error {
		switch status {
		case store.WatchStatusUnwatched:
			err := json.Unmarshal(movieJSON, &moviesByStatus.UnwatchedMovies)
			if err != nil {
				logger.Error(err.Error(), slog.String("marshalType", "unwatchedMovies"))
				return err
			}
		case store.WatchStatusSelected:
			// this will always be 1 movie, but we have an array from the agg so we'll just always take the first one
			selectedMovies := []*PartyMovie{}
			err := json.Unmarshal(movieJSON, &selectedMovies)
			if err != nil {
				logger.Error(err.Error(), slog.String("marshalType", "selectedMovie"))
				return err
			}
			if len(selectedMovies) > 0 {
				moviesByStatus.SelectedMovie = selectedMovies[0]
			}
		case store.WatchStatusWatched:
			err := json.Unmarshal(movieJSON, &moviesByStatus.WatchedMovies)
			if err != nil {
				logger.Error(err.Error(), slog.String("marshalType", "watchedMovies"))
				return err
			}
		}
		return nil
	})
	if err != nil {
		return MoviesByStatus{}, err
	}

	return moviesByStatus, nil
}

func (p Party) AddMovie(ctx context.Context, watcherID, idMovie int) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("Party").Start(ctx, "AddMovie")
	defer span.End()
	err := p.db.CreatePartyMovie(ctx, p.ID, idMovie, watcherID)
	if err != nil {
		return err
	}
	return nil
}

func (p Party) HasMovieAdded(ctx context.Context, movieID int) (bool, error) {
	exists, err := p.db.MovieAddedToParty(ctx, p.ID, movieID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *Party) GetPartyMembers(ctx context.Context) error {
	err := p.db.GetPartyMembers(ctx, p.ID, func(firstName, lastName string, id int, owner bool, joinedAt time.Time) {
		p.Members = append(p.Members, PartyMember{
			FirstName: firstName,
			LastName:  lastName,
			ID:        id,
			Owner:     owner,
			JoinedOn:  joinedAt,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

// generate a random 6 character string
func generateRandomString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
