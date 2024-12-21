package partymgmt

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jm96441n/movieswithfriends/store"
)

type moviesRepository interface {
	GetMovieByTMDBID(context.Context, int) (*store.Movie, error)
	CreateMovie(context.Context, *store.Movie) (*store.Movie, error)
}

type movieFetcher interface {
	Search(ctx context.Context, searchTerm string, page int) (SearchResults, error)
	GetMovie(ctx context.Context, tmdbID int) (*TMDBMovie, error)
	GetGenre(int) (Genre, error)
}

type MovieService struct {
	moviesRepository moviesRepository
	tmdbClient       movieFetcher
}

func NewMovieService(client *TMDBClient, moviesRepository moviesRepository) *MovieService {
	return &MovieService{
		tmdbClient:       client,
		moviesRepository: moviesRepository,
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

func (m *MovieService) CreateMovie(ctx context.Context, logger *slog.Logger, tmdbID int) (*store.Movie, error) {
	movie, err := m.moviesRepository.GetMovieByTMDBID(ctx, tmdbID)
	if movie != nil {
		logger.InfoContext(ctx, "movie found in db", slog.Any("movie", movie.Title))
		return movie, nil
	}

	if !errors.Is(err, store.ErrNoRecord) {
		logger.ErrorContext(ctx, "Failed to get movie by tmdbid", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return nil, err
	}

	err = nil

	tmdbMovie, err := m.tmdbClient.GetMovie(ctx, tmdbID)
	if err != nil || tmdbMovie == nil {
		logger.ErrorContext(ctx, "Failed to get movie from tmdb", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return nil, err
	}

	movie, err = m.moviesRepository.CreateMovie(ctx, tmdbMovie.ToStoreMovie())
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create movie", slog.Any("err", err), slog.Any("movie", tmdbMovie.Title))
		return nil, err
	}
	logger.InfoContext(ctx, "movie created in db", slog.Any("movie", movie))

	return movie, nil
}
