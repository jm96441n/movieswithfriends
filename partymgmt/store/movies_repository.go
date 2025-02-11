package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/metrics"
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
	Budget      *int
}

const (
	findMovieByIDQuery = `
  SELECT 
    id_movie, 
    title, 
    release_date, 
    overview, 
    tagline, 
    poster_url, 
    tmdb_id, 
    trailer_url,
    runtime,
    genres,
    budget
  FROM movies WHERE %s = $1`
)

type GetAssignFn func(*GetMovieResult)

// GetMovieByTMDBID returns a movie from the database by its TMDB ID
func (p *MoviesRepository) GetMovieByTMDBID(ctx context.Context, id int, assignFn GetAssignFn) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "MoviesRepository.GetMovieByTMDBID")
	defer span.End()
	query := fmt.Sprintf(findMovieByIDQuery, "tmdb_id")
	return p.getMovieBySomeID(ctx, id, assignFn, query)
}

// GetMovieByID returns a movie from the database by its ID
func (p *MoviesRepository) GetMovieByID(ctx context.Context, id int, assignFn GetAssignFn) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "MoviesRepository.GetMovieByID")
	defer span.End()
	query := fmt.Sprintf(findMovieByIDQuery, "id_movie")
	return p.getMovieBySomeID(ctx, id, assignFn, query)
}

func (p *MoviesRepository) getMovieBySomeID(ctx context.Context, id int, assignFn GetAssignFn, query string) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "MoviesRepository.getMovieBySomeID")
	defer span.End()
	row := p.db.QueryRow(ctx, query, id)

	res := &GetMovieResult{}
	var releaseDate time.Time
	err := row.Scan(&res.ID, &res.Title, &releaseDate, &res.Overview, &res.Tagline, &res.PosterURL, &res.TMDBID, &res.TrailerURL, &res.Runtime, &res.Genres, &res.Budget)
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
  genres,
  budget
  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id_movie`

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
	Budget      int
}

// CreateMovie creates a movie in the database
func (p *MoviesRepository) CreateMovie(ctx context.Context, createParams CreateMovieParams) (int, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "MoviesRepository.CreateMovie")
	defer span.End()
	var (
		releaseDate time.Time
		err         error
	)

	if createParams.ReleaseDate != "" {
		releaseDate, err = time.Parse("2006-01-02", createParams.ReleaseDate)
		if err != nil {
			return 0, err
		}
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
		createParams.Budget,
	).Scan(&movieID)
	if err != nil {
		return 0, err
	}

	return movieID, nil
}

const getMovieTMDBIDsFromPartyQuery = `select movies.tmdb_id from movies
join party_movies on movies.id_movie = party_movies.id_movie
where party_movies.id_party = $1 AND movies.tmdb_id = any($2);`

func (p *MoviesRepository) GetMovieTMDBIDsFromParty(ctx context.Context, partyID int, tmdbIDs []int, assignFn func(int)) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "MoviesRepository.GetMovieTMDBIDsFromParty")
	defer span.End()
	rows, err := p.db.Query(ctx, getMovieTMDBIDsFromPartyQuery, partyID, tmdbIDs)
	if err != nil {
		return err
	}

	var id int

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		assignFn(id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
