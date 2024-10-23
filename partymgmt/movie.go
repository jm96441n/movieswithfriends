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

type MovieService struct {
	logger           *slog.Logger
	moviesRepository moviesRepository
	tmdbClient       *TMDBClient
}

func NewMovieService(client *TMDBClient, logger *slog.Logger, moviesRepository moviesRepository) *MovieService {
	return &MovieService{
		tmdbClient:       client,
		logger:           logger,
		moviesRepository: moviesRepository,
	}
}

func (m *MovieService) SearchMovies(ctx context.Context, searchTerm string) ([]store.Movie, error) {
	result, err := m.tmdbClient.Search(ctx, searchTerm, 1)
	if err != nil {
		return nil, err
	}

	for idx := range result.Movies {
		result.Movies[idx].URL = fmt.Sprintf("/movies/%d", result.Movies[idx].TMDBID)
		result.Movies[idx].PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s", result.Movies[idx].PosterURL)
	}

	return result.Movies, nil
}

func (m *MovieService) CreateMovie(ctx context.Context, tmdbID int) (*store.Movie, error) {
	movie, err := m.moviesRepository.GetMovieByTMDBID(ctx, tmdbID)
	if movie != nil {
		m.logger.Info("movie found in db", slog.Any("movie", movie.Title))
		return movie, nil
	}

	if !errors.Is(err, store.ErrNoRecord) {
		m.logger.Error("Failed to get movie by tmdbid", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return nil, err
	}

	err = nil

	movie, err = m.tmdbClient.GetMovie(ctx, tmdbID)
	if err != nil {
		m.logger.Error("Failed to get movie from tmdb", slog.Any("err", err), slog.Any("tmdbID", tmdbID))
		return nil, err
	}

	movie, err = m.moviesRepository.CreateMovie(ctx, movie)
	if err != nil {
		m.logger.Error("Failed to create movie", slog.Any("err", err), slog.Any("movie", movie.Title))
		return nil, err
	}
	m.logger.Info("movie created in db", slog.Any("movie", movie))

	return movie, nil
}
