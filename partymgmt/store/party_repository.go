package store

import (
	"context"
	"errors"
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
    count(distinct party_movies.id_movie) as movie_count
  from parties
  join party_members on party_members.id_party = parties.id_party
  left join party_movies on party_movies.id_party = parties.id_party
  where parties.id_party = $1
  group by parties.id_party;
`

func (p *PartyRepository) GetPartyByIDWithStats(ctx context.Context, id int) (GetPartyByIDWithStatsResult, error) {
	// logger.Info("GetPartyByIDWithStats", "id", id)
	party := GetPartyByIDWithStatsResult{}
	err := p.db.QueryRow(ctx, getPartyByIDWithStatsQuery, id).Scan(&party.ID, &party.Name, &party.ShortID, &party.MemberCount, &party.MovieCount)
	if err != nil {
		return GetPartyByIDWithStatsResult{}, err
	}
	return party, nil
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

const getPartiesQueryForMovie = `
with filtered_party_movies as(select * from party_movies where id_movie = $1)select parties.id_party, parties.name, movies.id_movie is not null as is_movie
  from parties
  join party_members on party_members.id_party = parties.id_party
  left outer join filtered_party_movies on filtered_party_movies.id_party = parties.id_party 
  left outer join movies on filtered_party_movies.id_movie = movies.id_movie
  where party_members.id_member = $2;
`

func (p *PartyRepository) GetPartiesByMemberIDForCurrentMovie(ctx context.Context, idMovie int, idMember int) ([]Party, error) {
	rows, err := p.db.Query(ctx, getPartiesQueryForMovie, idMovie, idMember)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var parties []Party

	for rows.Next() {
		var party Party
		err := rows.Scan(&party.ID, &party.Name, &party.MovieAdded)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

const AddMovietoPartyQuery = `insert into party_movies (id_party, id_movie, id_added_by) values ($1, $2, $3)`

func (p *PartyRepository) AddMovieToParty(ctx context.Context, idParty, idMovie, id_added_by int) error {
	_, err := p.db.Exec(ctx, AddMovietoPartyQuery, idParty, idMovie, id_added_by)
	if err != nil {
		return err
	}
	return nil
}

const GetPartiesByMemberIDQuery = `
  with current_member_parties as (
    select 
      parties.id_party,
      parties.name
    from parties
    join party_members on party_members.id_party = parties.id_party
    where party_members.id_member = $1
  )
    select 
      current_member_parties.id_party,
      current_member_parties.name,
      count(distinct party_members.id_member) as member_count,
      count(distinct party_movies.id_movie) as movie_count
    from party_members
    left join party_movies on party_movies.id_party = party_members.id_party
    join current_member_parties on current_member_parties.id_party = party_members.id_party
    where party_members.id_party = current_member_parties.id_party
    group by current_member_parties.id_party, current_member_parties.name;
`

func (p *PartyRepository) MarkPartyMovieAsWatched(ctx context.Context, idParty, idMovie int) error {
	// pg.logger.Info("MarkMovieAsWatched", "idParty", idParty, "idMovie", idMovie)
	curTime := time.Now()
	return p.updatePartyMovieStatus(ctx, idParty, idMovie, WatchStatusWatched, &curTime)
}

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
)
  select id_movie 
  from party_movies 
  where id_party = $1 and id_added_by = (select id_member from selected_member_id) AND watch_status = 'unwatched'
  order by random()
  limit 1;
`

func (p *PartyRepository) SelectMovieForParty(ctx context.Context, idParty int) error {
	var selectedMovieID int
	err := p.db.QueryRow(ctx, selectMovieForPartyQuery, idParty).Scan(&selectedMovieID)
	if err != nil {
		return err
	}

	return p.updatePartyMovieStatus(ctx, idParty, selectedMovieID, WatchStatusSelected, nil)
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
