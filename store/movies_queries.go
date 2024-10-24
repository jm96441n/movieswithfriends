package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type FullName struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Movie struct {
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	Tagline     string `json:"tagline"`
	PosterURL   string `json:"poster_path"`
	TrailerURL  string `json:"trailer_url"`
	URL         string
	ID          int
	TMDBID      int `json:"id"`
	WatchStatus WatchStatusEnum
	AddedBy     FullName `json:"added_by"`
}

const (
	findMovieByTMDBIDQuery = `SELECT id_movie, title, release_date, overview, tagline, poster_url, tmdb_id, trailer_url FROM movies WHERE tmdb_id = $1`
	findMovieByIDQuery     = `SELECT id_movie, title, release_date, overview, tagline, poster_url, tmdb_id, trailer_url FROM movies WHERE id_movie = $1`
)

// GetMovieByTMDBID returns a movie from the database by its TMDB ID
func (p *PGStore) GetMovieByTMDBID(ctx context.Context, id int) (*Movie, error) {
	row := p.db.QueryRow(ctx, findMovieByTMDBIDQuery, id)

	movie := &Movie{}
	var releaseDate time.Time
	err := row.Scan(&movie.ID, &movie.Title, &releaseDate, &movie.Overview, &movie.Tagline, &movie.PosterURL, &movie.TMDBID, &movie.TrailerURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		}

		return nil, err
	}

	movie.ReleaseDate = releaseDate.Format("2006-01-02")

	return movie, nil
}

// GetMovieByID returns a movie from the database by its ID
func (p *PGStore) GetMovieByID(ctx context.Context, id int) (Movie, error) {
	row := p.db.QueryRow(ctx, findMovieByIDQuery, id)

	movie := Movie{}
	var releaseDate time.Time
	err := row.Scan(&movie.ID, &movie.Title, &releaseDate, &movie.Overview, &movie.Tagline, &movie.PosterURL, &movie.TMDBID, &movie.TrailerURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Movie{}, ErrNoRecord
		}

		return Movie{}, err
	}

	movie.ReleaseDate = releaseDate.Format("2006-01-02")

	return movie, nil
}

const insertMovieQuery = `INSERT INTO movies(title, release_date, overview, tagline, poster_url, tmdb_id, trailer_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id_movie`

// CreateMovie creates a movie in the database
func (p *PGStore) CreateMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
	if err != nil {
		return nil, err
	}

	err = p.db.QueryRow(ctx, insertMovieQuery, movie.Title, releaseDate, movie.Overview, movie.Tagline, movie.PosterURL, movie.TMDBID, movie.TrailerURL).Scan(&movie.ID)
	if err != nil {
		return nil, err
	}

	return movie, nil
}
