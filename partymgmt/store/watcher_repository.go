package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WatcherRepository struct {
	db *pgxpool.Pool
}

func NewWatcherRepository(db *pgxpool.Pool) *WatcherRepository {
	return &WatcherRepository{db: db}
}

type WatchedMoviesForWatcherResult struct {
	IDMovie   int
	Title     string
	WatchDate time.Time
	PartyName string
}

const getWatchedMoviesForWatcher = `
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

func (p *WatcherRepository) GetWatchedMoviesForWatcher(ctx context.Context, idWatcher int, offset int) ([]WatchedMoviesForWatcherResult, error) {
	rows, err := p.db.Query(ctx, getWatchedMoviesForWatcher, idWatcher, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var movies []WatchedMoviesForWatcherResult
	for rows.Next() {
		var movie WatchedMoviesForWatcherResult
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

const getWatchedMoviesCountForWatcher = `
  SELECT count(party_movies.*)
  FROM party_movies
  JOIN party_members ON party_members.id_party = party_movies.id_party
  WHERE party_members.id_member = $1 AND party_movies.watch_status = 'watched'
`

func (p *WatcherRepository) GetWatchedMoviesCountForMember(ctx context.Context, logger *slog.Logger, idMember int) (int, error) {
	var count int
	err := p.db.QueryRow(ctx, getWatchedMoviesCountForWatcher, idMember).Scan(&count)
	if err != nil {
		logger.Error("failed to get watched movies count", "error", err)
		return 0, err
	}

	return count, nil
}

const getPartiesForWatcherQuery = `
  with current_member_parties as (
    select 
      parties.id_party,
      parties.name,
      party_members.created_at
    from parties
    join party_members on party_members.id_party = parties.id_party
    where party_members.id_member = $1
  )
    select 
      current_member_parties.id_party,
      current_member_parties.name,
      current_member_parties.created_at,
      count(distinct party_members.id_member) as member_count,
      count(distinct party_movies.id_movie) as movie_count
    from party_members
    left join party_movies on party_members.id_party = party_movies.id_party
    join current_member_parties on current_member_parties.id_party = party_members.id_party
    where party_members.id_party = current_member_parties.id_party
    group by 
      current_member_parties.id_party, 
      current_member_parties.name, 
      current_member_parties.created_at
    order by current_member_parties.created_at desc  -- Order by created_on
    limit $2;
`

type PartiesForWatcherResult struct {
	ID          int
	Name        string
	MemberCount int
	MovieCount  int
}

func (p *WatcherRepository) GetPartiesForWatcher(ctx context.Context, watcherID, limit int) ([]PartiesForWatcherResult, error) {
	rows, err := p.db.Query(ctx, getPartiesForWatcherQuery, watcherID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []PartiesForWatcherResult
	for rows.Next() {
		var party PartiesForWatcherResult
		var t time.Time
		err := rows.Scan(&party.ID, &party.Name, &t, &party.MemberCount, &party.MovieCount)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return parties, nil
}

const getPartiesWithMovieQuery = `
  select parties.id_party, parties.name
  from parties
  join party_members on party_members.id_party = parties.id_party
  join party_movies on party_movies.id_party = parties.id_party 
  where party_movies.id_movie = $1 AND party_members.id_member = $2;
`

func (p *WatcherRepository) GetWatcherPartiesWithMovie(ctx context.Context, logger *slog.Logger, idMovie int, idMember int, assignFn func(int, string)) error {
	rows, err := p.db.Query(ctx, getPartiesWithMovieQuery, idMovie, idMember)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			partyID   int
			partyName string
		)
		err := rows.Scan(&partyID, &partyName)
		if err != nil {
			return err
		}

		assignFn(partyID, partyName)
	}
	return nil
}

const getPartiesWithoutMovieQuery = `
  select parties.id_party, parties.name
  from parties
  left join party_movies ON parties.id_party = party_movies.id_party AND party_movies.id_movie = $1
  join party_members on party_members.id_party = parties.id_party
  where pm.id_movie IS NULL AND party_members.id_member = $2;
`

func (p *WatcherRepository) GetWatcherPartiesWithoutMovie(ctx context.Context, logger *slog.Logger, idMovie int, idMember int, assignFn func(int, string)) error {
	rows, err := p.db.Query(ctx, getPartiesWithoutMovieQuery, idMovie, idMember)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			partyID   int
			partyName string
		)
		err := rows.Scan(&partyID, &partyName)
		if err != nil {
			return err
		}

		assignFn(partyID, partyName)
	}
	return nil
}

const isOwnerQuery = `
  select pm.owner
  from party_members pm
  where pm.id_member = $1 and pm.id_party = $2;
`

func (p *WatcherRepository) WatcherOwnsParty(ctx context.Context, idWatcher, idParty int) (bool, error) {
	var isOwner bool
	err := p.db.QueryRow(ctx, isOwnerQuery, idWatcher, idParty).Scan(&isOwner)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}
