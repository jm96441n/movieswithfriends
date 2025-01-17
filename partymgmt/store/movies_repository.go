package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MoviesRepository struct {
	db *pgxpool.Pool
}

func NewMoviesRepository(db *pgxpool.Pool) *MoviesRepository {
	return &MoviesRepository{db: db}
}

type FullName struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GetMovieResult struct {
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
}

const (
	findMovieByTMDBIDQuery = `SELECT id_movie, title, release_date, overview, tagline, poster_url, tmdb_id, trailer_url FROM movies WHERE tmdb_id = $1`
	findMovieByIDQuery     = `SELECT id_movie, title, release_date, overview, tagline, poster_url, tmdb_id, trailer_url FROM movies WHERE id_movie = $1`
)

type GetAssignFn func(*GetMovieResult)

// GetMovieByTMDBID returns a movie from the database by its TMDB ID
func (p *MoviesRepository) GetMovieByTMDBID(ctx context.Context, id int, assignFn GetAssignFn) error {
	row := p.db.QueryRow(ctx, findMovieByTMDBIDQuery, id)

	res := &GetMovieResult{}
	var releaseDate time.Time
	err := row.Scan(&res.ID, &res.Title, &releaseDate, &res.Overview, &res.Tagline, &res.PosterURL, &res.TMDBID, &res.TrailerURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoRecord
		}

		return err
	}

	res.ReleaseDate = releaseDate.Format("2006-01-02")
	assignFn(res)

	return nil
}

// GetMovieByID returns a movie from the database by its ID
func (p *MoviesRepository) GetMovieByID(ctx context.Context, id int, assignFn GetAssignFn) error {
	row := p.db.QueryRow(ctx, findMovieByIDQuery, id)

	res := &GetMovieResult{}
	var releaseDate time.Time
	err := row.Scan(&res.ID, &res.Title, &releaseDate, &res.Overview, &res.Tagline, &res.PosterURL, &res.TMDBID, &res.TrailerURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoRecord
		}

		return err
	}

	res.ReleaseDate = releaseDate.Format("2006-01-02")
	assignFn(res)

	return nil
}

const insertMovieQuery = `INSERT INTO movies(
  title, 
  release_date, 
  overview, 
  tagline, 
  poster_url, 
  tmdb_id, 
  trailer_url,
  rating,
  runtime,
  genres
  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id_movie`

type CreateMovieParams struct {
	Title       string
	ReleaseDate string
	Overview    string
	Tagline     string
	PosterURL   string
	TrailerURL  string
	URL         string
	Runtime     int
	Rating      float64
	Genres      []string
	GenreIDs    []int
	TMDBID      int
}

// CreateMovie creates a movie in the database
func (p *MoviesRepository) CreateMovie(ctx context.Context, createParams CreateMovieParams) (int, error) {
	releaseDate, err := time.Parse("2006-01-02", createParams.ReleaseDate)
	if err != nil {
		return 0, err
	}

	var movieID int

	err = p.db.QueryRow(ctx, insertMovieQuery,
		createParams.Title,
		releaseDate,
		createParams.Overview,
		createParams.Tagline,
		createParams.PosterURL,
		createParams.TMDBID,
		createParams.TrailerURL,
		createParams.Rating,
		createParams.Runtime,
		createParams.Genres,
	).Scan(&movieID)
	if err != nil {
		return 0, err
	}

	return movieID, nil
}
