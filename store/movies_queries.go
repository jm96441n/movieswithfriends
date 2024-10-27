package store

import (
	"context"
	"errors"
	"fmt"
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
	Runtime     int     `json:"runtime"`
	Rating      float64 `json:"vote_average"`
	Genres      []string
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

// CreateMovie creates a movie in the database
func (p *PGStore) CreateMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
	if err != nil {
		return nil, err
	}

	fmt.Println(movie)

	err = p.db.QueryRow(ctx, insertMovieQuery,
		movie.Title,
		releaseDate,
		movie.Overview,
		movie.Tagline,
		movie.PosterURL,
		movie.TMDBID,
		movie.TrailerURL,
		movie.Rating,
		movie.Runtime,
		movie.Genres,
	).Scan(&movie.ID)
	if err != nil {
		return nil, err
	}

	return movie, nil
}

const getMoviesForPartyQuery = `
SELECT *
FROM (
    SELECT 
      movies.id_movie,
      movies.title,
      movies.poster_url,
      profiles.first_name,
      profiles.last_name,
      party_movies.watch_status,
           ROW_NUMBER() OVER (PARTITION BY party_movies.watch_status ORDER BY party_movies.created_at DESC) as rn
    FROM movies
    INNER JOIN party_movies ON movies.id_movie = party_movies.id_movie
    JOIN profiles ON party_movies.id_added_by = profiles.id_profile
    WHERE party_movies.id_party = $1
) t
WHERE rn <= 10;
`

type MoviesByStatus struct {
	UnwatchedMovies []*Movie
	SelectedMovie   *Movie
	WatchedMovies   []*Movie
}

// GetMoviesForParty returns a paginated list of movies for a party grouped by watchStatus
func (p *PGStore) GetMoviesForParty(ctx context.Context, idParty, offset int) (MoviesByStatus, error) {
	movies := MoviesByStatus{
		UnwatchedMovies: []*Movie{},
		WatchedMovies:   []*Movie{},
	}
	rows, err := p.db.Query(ctx, getMoviesForPartyQuery, idParty)
	if err != nil {
		return MoviesByStatus{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			rn    int
			movie = &Movie{}
		)
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.PosterURL,
			&movie.AddedBy.FirstName,
			&movie.AddedBy.LastName,
			&movie.WatchStatus,
			&rn,
		)
		if err != nil {
			p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
			return MoviesByStatus{}, err
		}

		switch movie.WatchStatus {
		case WatchStatusUnwatched:
			movies.UnwatchedMovies = append(movies.UnwatchedMovies, movie)
		case WatchStatusSelected:
			movies.SelectedMovie = movie
		case WatchStatusWatched:
			movies.WatchedMovies = append(movies.WatchedMovies, movie)
		}
	}

	if err := rows.Err(); err != nil {
		p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
		return MoviesByStatus{}, err
	}

	p.logger.Info("GetPartyByIDWithMovies", "partyID", idParty)

	return movies, nil
}
