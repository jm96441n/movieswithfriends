package store

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartyRepository struct {
	db *pgxpool.Pool
}

func NewPartyRepository(db *pgxpool.Pool) *PartyRepository {
	return &PartyRepository{db: db}
}

type Party struct {
	ID              int
	Name            string
	ShortID         string
	MovieAdded      bool
	UnwatchedMovies []PartyMovie
	SelectedMovie   *PartyMovie
	WatchedMovies   []PartyMovie
}

type PartyMovie struct {
	ID          pgtype.Int8
	Title       pgtype.Text
	PosterURL   pgtype.Text
	WatchStatus pgtype.Text
}

const getPartyByIDQuery = `select id_party, name, short_id from parties where id_party = $1`

type GetPartyResult struct {
	ID      int
	Name    string
	ShortID string
}

func (p *PartyRepository) GetPartyByID(ctx context.Context, id int) (GetPartyResult, error) {
	res := GetPartyResult{}
	err := p.db.QueryRow(ctx, getPartyByIDQuery, id).Scan(&res.ID, &res.Name, &res.ShortID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetPartyResult{}, ErrNoRecord
		}
		return GetPartyResult{}, err
	}
	return res, nil
}

const getPartyByShortIDQuery = `select id_party, name, short_id from parties where short_id = $1`

func (p *PartyRepository) GetPartyByShortID(ctx context.Context, shortID string) (GetPartyResult, error) {
	party := GetPartyResult{}

	err := p.db.QueryRow(ctx, getPartyByShortIDQuery, shortID).Scan(&party.ID, &party.Name, &party.ShortID)
	if err != nil {
		return GetPartyResult{}, err
	}

	return party, nil
}

type GetPartyByIDWithStatsResult struct {
	ID           int
	Name         string
	ShortID      string
	MemberCount  int
	MovieCount   int
	WatchedCount int
}

const getPartyByIDWithStatsQuery = `
select
    parties.id_party,
    parties.name,
    parties.short_id,
    count(distinct party_members.id_member) as member_count,
    count(distinct party_movies.id_movie) as movie_count,
	  count(distinct case when party_movies.watch_status = 'watched' then party_movies.id_movie end) as watched_count
  from parties
  join party_members on party_members.id_party = parties.id_party
  left join party_movies on party_movies.id_party = parties.id_party
  where parties.id_party = $1
  group by parties.id_party;
`

type getAssignFn func(GetPartyByIDWithStatsResult)

func (p *PartyRepository) GetPartyByIDWithStats(ctx context.Context, id int, assignFn getAssignFn) error {
	// logger.Info("GetPartyByIDWithStats", "id", id)
	result := GetPartyByIDWithStatsResult{}
	err := p.db.QueryRow(ctx, getPartyByIDWithStatsQuery, id).
		Scan(
			&result.ID,
			&result.Name,
			&result.ShortID,
			&result.MemberCount,
			&result.MovieCount,
			&result.WatchedCount,
		)
	if err != nil {
		return err
	}

	assignFn(result)
	return nil
}

const (
	createPartyQuery              = `INSERT INTO parties (name, short_id) VALUES ($1, $2) returning id_party;`
	createPartyMemberAsOwnerQuery = `INSERT INTO party_members (id_member, id_party, owner) VALUES ($1, $2, true);`
)

func (p *PartyRepository) CreateParty(ctx context.Context, idMember int, name, shortID string) (int, error) {
	txn, err := p.db.Begin(ctx)
	if err != nil {
		return 0, err
	}

	defer txn.Rollback(ctx)

	var id int

	err = txn.QueryRow(ctx, createPartyQuery, name, shortID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgUniqueViolationCode {
				if pgErr.ConstraintName == "unique_parties_short_id" {
					return 0, ErrDuplicatePartyShortID
				}
			}
		}
		return 0, err
	}

	_, err = txn.Exec(ctx, createPartyMemberAsOwnerQuery, idMember, id)
	if err != nil {
		return 0, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

const AddMovietoPartyQuery = `insert into party_movies (id_party, id_movie, id_added_by) values ($1, $2, $3)`

func (p *PartyRepository) AddMovieToParty(ctx context.Context, idParty, idMovie, id_added_by int) error {
	_, err := p.db.Exec(ctx, AddMovietoPartyQuery, idParty, idMovie, id_added_by)
	if err != nil {
		return err
	}
	return nil
}

func (p *PartyRepository) MarkPartyMovieAsWatched(ctx context.Context, idParty, idMovie int) error {
	curTime := time.Now().UTC()
	return p.updatePartyMovieStatus(ctx, idParty, idMovie, WatchStatusWatched, &curTime)
}

const setCurrentSelectMoviesToUnwatched = `
UPDATE party_movies
SET watch_status = 'unwatched'
WHERE id_party = $1 AND watch_status = 'selected';
`

const selectMovieForPartyQuery = `
WITH party_members_for_selection AS (
  select distinct(id_added_by) as id_member
  from party_movies
  where id_party = $1
), selected_member_id AS (
  select id_member 
  from party_members_for_selection
  order by random()
  limit 1
), selected_movie AS (
  select id_movie 
  from party_movies 
  where id_party = $1 
  and id_added_by = (select id_member from selected_member_id) 
  AND watch_status = 'unwatched'
  order by random()
  limit 1
)
UPDATE party_movies
SET watch_status = 'selected'
WHERE id_movie = (select id_movie from selected_movie)
AND id_party = $1
RETURNING id_movie;
`

func (p *PartyRepository) SelectMovieForParty(ctx context.Context, idParty int) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, setCurrentSelectMoviesToUnwatched, idParty)
	if err != nil {
		return err
	}

	var selectedMovieID int
	err = tx.QueryRow(ctx, selectMovieForPartyQuery, idParty).Scan(&selectedMovieID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

const updateWatchStatusQuery = `
  update party_movies
  set watch_status = $1, watch_date = $4
  where id_party = $2 and id_movie = $3
`

func (p *PartyRepository) updatePartyMovieStatus(ctx context.Context, idParty, idMovie int, status WatchStatusEnum, watchDate *time.Time) error {
	_, err := p.db.Exec(ctx, updateWatchStatusQuery, status, idParty, idMovie, watchDate)
	if err != nil {
		return err
	}
	return nil
}

const AddFriendToPartyQuery = `insert into party_members (id_member, id_party) values($1, $2)`

func (p *PartyRepository) CreatePartyMember(ctx context.Context, idMember, idParty int) error {
	_, err := p.db.Exec(ctx, AddFriendToPartyQuery, idMember, idParty)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgUniqueViolationCode {
				return ErrMemberPartyCombinationNotUnique
			}
		}
		return err
	}
	return nil
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
func (p *PartyRepository) GetMoviesForParty(ctx context.Context, logger *slog.Logger, idParty, offset int, assignFn func(WatchStatusEnum, []byte) error) error {
	rows, err := p.db.Query(ctx, getMoviesForPartyQuery, idParty)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			status    WatchStatusEnum
			movieJSON []byte
		)
		err := rows.Scan(&status, &movieJSON)
		if err != nil {
			logger.Error(err.Error(), "query", getMoviesForPartyQuery)
			return err
		}

		err = assignFn(status, movieJSON)
		if err != nil {
			return err
		}

	}

	if err := rows.Err(); err != nil {
		logger.Error(err.Error(), "query", getMoviesForPartyQuery)
		return err
	}

	return nil
}

const createPartyMovieQuery = `insert into party_movies (id_party, id_movie, id_added_by) values ($1, $2, $3)`

// CreatePartMovie creates a movie within a party
func (p *PartyRepository) CreatePartyMovie(ctx context.Context, idParty, idMovie, idAddedBy int) error {
	_, err := p.db.Exec(ctx, createPartyMovieQuery, idParty, idMovie, idAddedBy)
	if err != nil {
		return err
	}
	return nil
}

const movieInPartyQuery = `SELECT EXISTS(select 1 from party_movies where id_movie = $1 and id_party = $2)`

func (p *PartyRepository) MovieAddedToParty(ctx context.Context, idParty, idMovie int) (bool, error) {
	var exists bool

	err := p.db.QueryRow(ctx, movieInPartyQuery, idMovie, idParty).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
