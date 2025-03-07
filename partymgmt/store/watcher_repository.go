package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jm96441n/movieswithfriends/metrics"
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
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatchedMoviesForWatcher")
	defer span.End()
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
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatchedMoviesCountForMember")
	defer span.End()
	var count int
	err := p.db.QueryRow(ctx, getWatchedMoviesCountForWatcher, idMember).Scan(&count)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get watched movies count", "error", err)
		return 0, err
	}

	return count, nil
}

const getPartiesForWatcherQuery = `
  with current_member_parties as (
    select 
      parties.id_party,
      parties.name,
      parties.created_at,
      parties.id_owner
    from parties
    join party_members on party_members.id_party = parties.id_party
    where party_members.id_member = $1
  )
    select 
      current_member_parties.id_party,
      current_member_parties.name,
      current_member_parties.created_at,
      current_member_parties.id_owner,
      count(distinct party_members.id_member) as member_count,
      count(distinct party_movies.id_movie) as movie_count
    from party_members
    left join party_movies on party_members.id_party = party_movies.id_party
    join current_member_parties on current_member_parties.id_party = party_members.id_party
    where party_members.id_party = current_member_parties.id_party
    group by 
      current_member_parties.id_party, 
      current_member_parties.name, 
      current_member_parties.created_at,
      current_member_parties.id_owner
    order by current_member_parties.created_at desc  -- Order by created_on
    limit $2;
`

type assignPartyFn func(ctx context.Context, id int, name string, movieCount int, memberCount int, idOwner int)

func (p *WatcherRepository) GetPartiesForWatcher(ctx context.Context, watcherID, limit int, assignFn assignPartyFn) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetPartiesForWatcher")
	defer span.End()

	rows, err := p.db.Query(ctx, getPartiesForWatcherQuery, watcherID, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int
			name        string
			memberCount int
			movieCount  int
			t           time.Time
			idOwner     int
		)
		err := rows.Scan(&id, &name, &t, &idOwner, &memberCount, &movieCount)
		if err != nil {
			return err
		}
		assignFn(ctx, id, name, memberCount, movieCount, idOwner)
	}

	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

const getInvitedPartiesForWatcherQuery = `
  with current_member_parties as (
    select 
      parties.id_party,
      parties.name,
      parties.created_at,
      parties.id_owner
    from parties
    join invitations on invitations.id_party = parties.id_party
    where invitations.id_profile = $1
  )
    select 
      current_member_parties.id_party,
      current_member_parties.name,
      current_member_parties.created_at,
      current_member_parties.id_owner,
      count(distinct party_members.id_member) as member_count,
      count(distinct party_movies.id_movie) as movie_count
    from party_members
    left join party_movies on party_members.id_party = party_movies.id_party
    join current_member_parties on current_member_parties.id_party = party_members.id_party
    where party_members.id_party = current_member_parties.id_party
    group by 
      current_member_parties.id_party, 
      current_member_parties.name, 
      current_member_parties.created_at,
      current_member_parties.id_owner
    order by current_member_parties.created_at desc  -- Order by created_on
    limit $2;
`

func (p *WatcherRepository) GetInvitedPartiesForWatcher(ctx context.Context, watcherID, limit int, assignFn assignPartyFn) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetInvitedPartiesForWatcher")
	defer span.End()

	rows, err := p.db.Query(ctx, getInvitedPartiesForWatcherQuery, watcherID, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int
			name        string
			memberCount int
			movieCount  int
			idOwner     int
			t           time.Time
		)
		err := rows.Scan(&id, &name, &t, &idOwner, &memberCount, &movieCount)
		if err != nil {
			return err
		}
		assignFn(ctx, id, name, memberCount, movieCount, idOwner)
	}

	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

const getPartiesWithMovieQuery = `
  select parties.id_party, parties.name
  from parties
  join party_members on party_members.id_party = parties.id_party
  join party_movies on party_movies.id_party = parties.id_party 
  where party_movies.id_movie = $1 AND party_members.id_member = $2;
`

func (p *WatcherRepository) GetWatcherPartiesWithMovie(ctx context.Context, logger *slog.Logger, idMember int, idMovie int, assignFn func(int, string)) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatcherPartiesWithMovie")
	defer span.End()

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
  SELECT parties.id_party, parties.name, COALESCE(COUNT(pm2.id_movie), 0)
  FROM parties
  LEFT JOIN party_movies ON parties.id_party = party_movies.id_party AND party_movies.id_movie = $1
  LEFT JOIN party_movies pm2 ON parties.id_party = pm2.id_party
  JOIN party_members ON party_members.id_party = parties.id_party
  WHERE party_movies.id_movie IS NULL AND party_members.id_member = $2
  GROUP BY parties.id_party, parties.name;
`

func (p *WatcherRepository) GetWatcherPartiesWithoutMovie(ctx context.Context, logger *slog.Logger, idMember int, idMovie int, assignFn func(int, string, int)) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatcherPartiesWithoutMovie")
	defer span.End()
	rows, err := p.db.Query(ctx, getPartiesWithoutMovieQuery, idMovie, idMember)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			partyID    int
			partyName  string
			movieCount int
		)
		err := rows.Scan(&partyID, &partyName, &movieCount)
		if err != nil {
			return err
		}

		assignFn(partyID, partyName, movieCount)
	}
	return nil
}

const getPartiesWithMovieQueryByTMDBID = `
  select parties.id_party, parties.name
  from parties
  join party_members on party_members.id_party = parties.id_party
  join party_movies on party_movies.id_party = parties.id_party 
  join movies on party_movies.id_movie = movies.id_movie
  where movies.tmdb_id = $1 AND party_members.id_member = $2;
`

func (p *WatcherRepository) GetWatcherPartiesWithMovieByTMDBID(ctx context.Context, logger *slog.Logger, idMember int, tmdbID int, assignFn func(int, string)) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatcherPartiesWithMovieByTMDBID")
	defer span.End()
	logger.InfoContext(ctx, "args", slog.Any("tmdbID", tmdbID), slog.Any("idMember", idMember))
	rows, err := p.db.Query(ctx, getPartiesWithMovieQueryByTMDBID, tmdbID, idMember)
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

	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

const getPartiesWithoutMovieQueryByTMDBID = `
SELECT parties.id_party, parties.name, COALESCE(COUNT(pm2.id_movie), 0)
FROM parties
LEFT JOIN party_movies ON parties.id_party = party_movies.id_party
LEFT JOIN movies ON party_movies.id_movie = movies.id_movie AND movies.tmdb_id = $1
LEFT JOIN party_movies pm2 ON parties.id_party = pm2.id_party
JOIN party_members ON party_members.id_party = parties.id_party
WHERE movies.id_movie IS NULL AND party_members.id_member = $2
GROUP BY parties.id_party, parties.name;
`

func (p *WatcherRepository) GetWatcherPartiesWithoutMovieByTMDBID(ctx context.Context, logger *slog.Logger, idMember int, tmdbID int, assignFn func(int, string, int)) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatcherPartiesWithoutMovieByTMDBID")
	defer span.End()
	rows, err := p.db.Query(ctx, getPartiesWithoutMovieQueryByTMDBID, tmdbID, idMember)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			partyID    int
			partyName  string
			movieCount int
		)
		err := rows.Scan(&partyID, &partyName, &movieCount)
		if err != nil {
			return err
		}

		assignFn(partyID, partyName, movieCount)
	}
	return nil
}

const isOwnerQuery = `
  select pm.owner
  from party_members pm
  where pm.id_member = $1 and pm.id_party = $2;
`

func (p *WatcherRepository) WatcherOwnsParty(ctx context.Context, idWatcher, idParty int) (bool, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.WatcherOwnsParty")
	defer span.End()
	var isOwner bool
	err := p.db.QueryRow(ctx, isOwnerQuery, idWatcher, idParty).Scan(&isOwner)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

const getWatcherByEmailQuery = `
  SELECT p.id_profile
  FROM profiles p
  JOIN accounts a ON a.id_account = p.id_account
  WHERE a.email = $1;
`

type assignIDFn func(int)

func (p WatcherRepository) GetWatcherByEmail(ctx context.Context, email string, assignFn assignIDFn) error {
	ctx, span, _ := metrics.SpanFromContext(ctx, "WatcherRepository.GetWatcherByEmail")
	defer span.End()
	var id int
	err := p.db.QueryRow(ctx, getWatcherByEmailQuery, email).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoRecord
		}

		return err
	}

	assignFn(id)

	return nil
}
