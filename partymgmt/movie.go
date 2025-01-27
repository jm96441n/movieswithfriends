package partymgmt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type FullName struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Movie struct {
	ID          int
	Title       string
	ReleaseDate string
	Overview    string
	Tagline     string
	PosterURL   string
	TrailerURL  string
	Runtime     int
	Rating      float64
	Genres      []string
	TMDBID      int
	AddedBy     FullName
	Budget      int
}

type movieFetcher interface {
	Search(ctx context.Context, searchTerm string, page int) (SearchResults, error)
	GetMovie(ctx context.Context, tmdbID int) (*TMDBMovie, error)
	GetGenre(int) (Genre, error)
}

type MovieService struct {
	db         *store.MoviesRepository
	tmdbClient movieFetcher
}

func NewMovieService(client *TMDBClient, moviesRepository *store.MoviesRepository) *MovieService {
	return &MovieService{
		tmdbClient: client,
		db:         moviesRepository,
	}
}

func (m *MovieService) SearchMovies(ctx context.Context, logger *slog.Logger, searchTerm string) ([]TMDBMovie, error) {
	result, err := m.tmdbClient.Search(ctx, searchTerm, 1)
	if err != nil {
		return nil, err
	}

	for idx := range result.Movies {
		result.Movies[idx].URL = fmt.Sprintf("/movies/%d", result.Movies[idx].TMDBID)
		if result.Movies[idx].PosterURL != "" {
			result.Movies[idx].PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s", result.Movies[idx].PosterURL)
		} else {
			result.Movies[idx].PosterURL = "https://placehold.co/270x400?text=No+Poster+Available"
		}
		result.Movies[idx].Genres = make([]Genre, 0, len(result.Movies[idx].GenreIDs))
		for _, genreID := range result.Movies[idx].GenreIDs {
			genre, err := m.tmdbClient.GetGenre(genreID)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to get genre", slog.Any("err", err), slog.Any("genreID", genreID))
				continue
			}
			result.Movies[idx].Genres = append(result.Movies[idx].Genres, genre)
		}
	}

	return result.Movies, nil
}

func (m *MovieService) GetMovieTMDBIDsFromCurrentParty(ctx context.Context, logger *slog.Logger, partyID int, movies []TMDBMovie) (map[int]struct{}, error) {
	tmdbIDs := make([]int, 0, len(movies))
	for _, movie := range movies {
		tmdbIDs = append(tmdbIDs, movie.TMDBID)
	}

	movieIDSet := make(map[int]struct{})
	err := m.db.GetMovieTMDBIDsFromParty(ctx, partyID, tmdbIDs, func(id int) {
		movieIDSet[id] = struct{}{}
	})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get movie tmdbids from party", slog.Any("err", err), slog.Any("partyID", partyID))
		return nil, err
	}

	return movieIDSet, nil
}

type MovieID struct {
	TMDBID  *int
	MovieID *int
}

func (m MovieID) validate() error {
	if m.TMDBID == nil && m.MovieID == nil {
		return errors.New("TMDBID or MovieID must be set")
	}

	if m.TMDBID != nil && m.MovieID != nil {
		return errors.New("only one of TMDBID or MovieID must be set")
	}

	return nil
}

var ErrMovieDoesNotExist = errors.New("Movie cannot be found")

func (m *MovieService) GetMovie(ctx context.Context, logger *slog.Logger, movieID MovieID) (Movie, error) {
	movie := Movie{}
	err := movieID.validate()
	if err != nil {
		return Movie{}, err
	}

	switch {
	case movieID.TMDBID != nil:
		err = m.db.GetMovieByTMDBID(ctx, *movieID.TMDBID, convertGetResultToMovie(&movie))
	case movieID.MovieID != nil:
		err = m.db.GetMovieByID(ctx, *movieID.MovieID, convertGetResultToMovie(&movie))
	}

	if errors.Is(err, store.ErrNoRecord) {
		return Movie{}, fmt.Errorf("%w: %s", ErrMovieDoesNotExist, err)
	}

	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func (m *MovieService) GetOrCreateMovie(ctx context.Context, logger *slog.Logger, movieID MovieID) (int, error) {
	err := movieID.validate()
	if err != nil {
		return 0, err
	}

	// if we have the actual id of the movie then we know it exists
	if movieID.MovieID != nil {
		return *movieID.MovieID, nil
	}

	// this is racy, it's very possible that we duplicate the create movie path
	// this is fine for now
	movie, err := m.GetMovie(ctx, logger, movieID)
	if errors.Is(err, ErrMovieDoesNotExist) {
		// we know tmdbid is set from the validate call earlier
		id, err := m.CreateMovie(ctx, logger, *movieID.TMDBID)
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	if err != nil {
		return 0, err
	}

	// movie exists so we just return it
	return movie.ID, nil
}

func convertGetResultToMovie(movie *Movie) store.GetAssignFn {
	return func(res *store.GetMovieResult) {
		movie.ID = res.ID
		movie.Title = res.Title
		movie.ReleaseDate = res.ReleaseDate
		movie.Overview = res.Overview
		movie.Tagline = res.Tagline
		movie.PosterURL = res.PosterURL
		movie.TrailerURL = res.TrailerURL
		movie.Runtime = res.Runtime
		movie.Rating = res.Rating
		movie.Genres = res.Genres
		movie.TMDBID = res.TMDBID
		movie.Budget = res.Budget
	}
}

func (m *MovieService) CreateMovie(ctx context.Context, logger *slog.Logger, tmdbID int) (int, error) {
	movie := Movie{}
	err := m.db.GetMovieByTMDBID(ctx, tmdbID, convertGetResultToMovie(&movie))
	// zero value means it's been set so movie exists
	if movie.ID > 0 {
		logger.InfoContext(ctx, "movie found in db", slog.Any("movie", movie.Title))
		return movie.ID, nil
	}

	if !errors.Is(err, store.ErrNoRecord) {
		logger.ErrorContext(ctx, "Failed to get movie by tmdbid", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return 0, err
	}

	err = nil

	tmdbMovie, err := m.tmdbClient.GetMovie(ctx, tmdbID)
	if err != nil || tmdbMovie == nil {
		logger.ErrorContext(ctx, "Failed to get movie from tmdb", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return 0, err
	}

	fmt.Printf("%#v\n", tmdbMovie)

	movieID, err := m.db.CreateMovie(ctx, tmdbMovie.ToStoreMovie())
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create movie", slog.Any("err", err), slog.Any("movie", tmdbMovie.Title))
		return 0, err
	}
	logger.InfoContext(ctx, "movie created in db", slog.Any("movieID", movieID))

	return movieID, nil
}
