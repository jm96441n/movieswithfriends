package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
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

type UnwatchedMovie struct {
	ID          int       `json:"id_movie"`
	Title       string    `json:"title"`
	ReleaseDate string    `json:"release_date"`
	Rating      float64   `json:"vote_average"`
	Genres      []string  `json:"genres"`
	AddedBy     FullName  `json:"added_by"`
	AddedOnDate time.Time `json:"added_on_date"`
}

type WatchedMovie struct {
	ID        int       `json:"id_movie"`
	Title     string    `json:"title"`
	WatchDate time.Time `json:"watch_date"`
}

type SelectedMovie struct {
	ID          int      `json:"id_movie"`
	Title       string   `json:"title"`
	ReleaseDate string   `json:"release_date"`
	TrailerURL  string   `json:"trailer_url"`
	PosterURL   string   `json:"poster_url"`
	Runtime     int      `json:"runtime"`
	Rating      float64  `json:"vote_average"`
	Tagline     string   `json:"tagline"`
	Genres      []string `json:"genres"`
	AddedBy     FullName `json:"added_by"`
}

type MoviesByStatus struct {
	UnwatchedMovies []*UnwatchedMovie
	SelectedMovie   *SelectedMovie
	WatchedMovies   []*WatchedMovie
}

const getMoviesForPartyQuery = `
SELECT watch_status, jsonb_agg(
  jsonb_build_object(
        'id_movie', id_movie,
        'title', title,
        'poster_url', poster_url,
        'added_by', jsonb_build_object(
            'first_name', first_name,
            'last_name', last_name
        ),
        'watch_date', watch_date,
        'genres', genres
    )
) as movies
FROM (
    SELECT *
    FROM (
        SELECT 
          movies.id_movie,
          movies.title,
          movies.poster_url,
          movies.genres,
          profiles.first_name,
          profiles.last_name,
          party_movies.watch_status,
          party_movies.watch_date,
          ROW_NUMBER() OVER (PARTITION BY party_movies.watch_status ORDER BY party_movies.created_at DESC) as rn
        FROM movies
        INNER JOIN party_movies ON movies.id_movie = party_movies.id_movie
        JOIN profiles ON party_movies.id_added_by = profiles.id_profile
        WHERE party_movies.id_party = $1
    ) t
    WHERE rn <= 10
) movie_data
GROUP BY watch_status;
`

// GetMoviesForParty returns a paginated list of movies for a party grouped by watchStatus
func (p *PGStore) GetMoviesForParty(ctx context.Context, idParty, offset int) (MoviesByStatus, error) {
	movies := MoviesByStatus{
		UnwatchedMovies: []*UnwatchedMovie{},
		WatchedMovies:   []*WatchedMovie{},
	}
	rows, err := p.db.Query(ctx, getMoviesForPartyQuery, idParty)
	if err != nil {
		return MoviesByStatus{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			status    WatchStatusEnum
			movieJSON []byte
		)
		err := rows.Scan(&status, &movieJSON)
		if err != nil {
			p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
			return MoviesByStatus{}, err
		}

		switch status {
		case WatchStatusUnwatched:
			unwatchedMovies := []*UnwatchedMovie{}
			err = json.Unmarshal(movieJSON, &unwatchedMovies)
			if err != nil {
				p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
				return MoviesByStatus{}, err
			}
			movies.UnwatchedMovies = unwatchedMovies
		case WatchStatusSelected:
			// this will always be 1 movie, but we have an array from the agg so we'll just always take the first one
			selectedMovies := []*SelectedMovie{}
			err = json.Unmarshal(movieJSON, &selectedMovies)
			if err != nil {
				p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
				return MoviesByStatus{}, err
			}
			if len(selectedMovies) > 0 {
				movies.SelectedMovie = selectedMovies[0]
			}
		case WatchStatusWatched:
			watchedMovies := []*WatchedMovie{}
			err = json.Unmarshal(movieJSON, &watchedMovies)
			if err != nil {
				p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
				return MoviesByStatus{}, err
			}
			movies.WatchedMovies = watchedMovies
		}
	}

	if err := rows.Err(); err != nil {
		p.logger.Error(err.Error(), "query", getMoviesForPartyQuery)
		return MoviesByStatus{}, err
	}

	return movies, nil
}
