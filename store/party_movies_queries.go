package store

import (
	"context"
	"time"
)

type WatchedMoviesForMemberResult struct {
	IDMovie   int
	Title     string
	WatchDate time.Time
	PartyName string
}

const getWatchedMoviesForMember = `
  SELECT
    movies.id_movie,
    movies.title,
    party_movies.watch_date ,
    parties.name
  FROM party_movies
  JOIN movies ON movies.id_movie = party_movies.id_movie
  JOIN party_members ON party_members.id_party = party_movies.id_party
  JOIN parties ON parties.id_party = party_movies.id_party 
  WHERE party_members.id_member = $1 AND party_movies.watch_status = 'watched'
  ORDER BY party_movies.watch_date DESC
  LIMIT 5
  OFFSET $2;
`

func (p *PGStore) GetWatchedMoviesForMember(ctx context.Context, idMember int, offset int) ([]WatchedMoviesForMemberResult, error) {
	rows, err := p.db.Query(ctx, getWatchedMoviesForMember, idMember, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var movies []WatchedMoviesForMemberResult
	for rows.Next() {
		var movie WatchedMoviesForMemberResult
		err := rows.Scan(&movie.IDMovie, &movie.Title, &movie.WatchDate, &movie.PartyName)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return movies, nil
}

const getWatchedMoviesCountForMember = `
  SELECT count(party_movies.*)
  FROM party_movies
  JOIN party_members ON party_members.id_party = party_movies.id_party
  WHERE party_members.id_member = $1 AND party_movies.watch_status = 'watched'
`

func (p *PGStore) GetWatchedMoviesCountForMember(ctx context.Context, idMember int) (int, error) {
	var count int
	err := p.db.QueryRow(ctx, getWatchedMoviesCountForMember, idMember).Scan(&count)
	if err != nil {
		p.logger.Error("failed to get watched movies count", "error", err)
		return 0, err
	}

	return count, nil
}
