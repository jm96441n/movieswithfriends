package partymgmt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

var ErrMemberExistsInParty = errors.New("member already exists in party")

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PartyService struct {
	logger *slog.Logger
	db     store.PartyRepository
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
	IDWatcher int
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
	IDOwner      int

	MoviesByStatus MoviesByStatus
	db             store.PartyRepository
}

func NewPartyService(logger *slog.Logger, db store.PartyRepository) PartyService {
	return PartyService{
		logger: logger,
		db:     db,
	}
}

// TODO: group these args so they can't be mixed up
func (s PartyService) NewParty(ctx context.Context, id int, name string, movieCount, memberCount, idOwner int) Party {
	_, span, _ := metrics.SpanFromContext(ctx, "PartyService.NewParty")
	defer span.End()
	return Party{
		ID:          id,
		Name:        name,
		MovieCount:  movieCount,
		MemberCount: memberCount,
		IDOwner:     idOwner,
		db:          s.db,
	}
}

func (s PartyService) CreateParty(ctx context.Context, idMember int, name string) (int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "PartyService.CreateParty")
	defer span.End()
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

func (s PartyService) GetPartyWithMovies(ctx context.Context, logger *slog.Logger, id int) (Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "PartyService.GetPartyWithMovies")
	defer span.End()

	var party Party
	err := s.db.GetPartyByIDWithStats(ctx, id, func(id int, name string, ownerID int, memberCount, movieCount, watchedCount int) {
		party = s.NewParty(ctx, id, name, movieCount, memberCount, ownerID)
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get party by id", slog.Any("error", err))
		return Party{}, err
	}

	err = party.GetPartyMembers(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get party members", slog.Any("error", err))
		return Party{}, err
	}

	moviesByStatus, err := party.GetMoviesByStatus(ctx, logger)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get movies for party", slog.Any("error", err))
		return Party{}, err
	}

	party.MoviesByStatus = moviesByStatus

	return party, nil
}

func (s PartyService) GetPartyByShortID(ctx context.Context, shortID string) (Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "PartyService.GetPartyByShortID")
	defer span.End()

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

func (p Party) AcceptInvite(ctx context.Context, logger *slog.Logger, watcherID int) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Party.AddMember")
	defer span.End()

	err := p.db.RunInTransaction(ctx, func(ctx context.Context, db store.PartyRepository) error {
		err := db.DeleteInvite(ctx, watcherID, p.ID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to delete invite", slog.Any("watcher_id", watcherID), slog.Any("party_id", p.ID))
			return err
		}

		err = db.CreatePartyMember(ctx, watcherID, p.ID)

		// if the party member exists then it's fine because it's all done
		if errors.Is(err, store.ErrMemberPartyCombinationNotUnique) {
			logger.DebugContext(ctx, "watcher already in party", slog.Any("watcher_id", watcherID), slog.Any("party_id", p.ID))
			return nil
		}

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p Party) GetMoviesByStatus(ctx context.Context, logger *slog.Logger) (MoviesByStatus, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Party.GetMoviesByStatus")
	defer span.End()

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
				logger.ErrorContext(ctx, err.Error(), slog.String("marshalType", "unwatchedMovies"))
				return err
			}
		case store.WatchStatusSelected:
			// this will always be 1 movie, but we have an array from the agg so we'll just always take the first one
			selectedMovies := []*PartyMovie{}
			err := json.Unmarshal(movieJSON, &selectedMovies)
			if err != nil {
				logger.ErrorContext(ctx, err.Error(), slog.String("marshalType", "selectedMovie"))
				return err
			}
			if len(selectedMovies) > 0 {
				moviesByStatus.SelectedMovie = selectedMovies[0]
			}
		case store.WatchStatusWatched:
			err := json.Unmarshal(movieJSON, &moviesByStatus.WatchedMovies)
			if err != nil {
				logger.ErrorContext(ctx, err.Error(), slog.String("marshalType", "watchedMovies"))
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
	ctx, span, _ := metrics.SpanFromContext(ctx, "Party.AddMovie")
	defer span.End()

	err := p.db.CreatePartyMovie(ctx, p.ID, idMovie, watcherID)
	if err != nil {
		return err
	}
	return nil
}

func (p Party) HasMovieAdded(ctx context.Context, movieID int) (bool, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Party.HasMovieAdded")
	defer span.End()

	exists, err := p.db.MovieAddedToParty(ctx, p.ID, movieID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *Party) GetPartyMembers(ctx context.Context) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "Party.GetPartyMembers")
	defer span.End()

	err := p.db.GetPartyMembers(ctx, p.ID, func(firstName, lastName string, id int, joinedAt time.Time, idWatcher int) {
		p.Members = append(p.Members, PartyMember{
			FirstName: firstName,
			LastName:  lastName,
			ID:        id,
			JoinedOn:  joinedAt,
			IDWatcher: idWatcher,
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
