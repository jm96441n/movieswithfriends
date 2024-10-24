package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type Party struct {
	ID              int
	Name            string
	ShortID         string
	MovieAdded      bool
	UnwatchedMovies []Movie
	SelectedMovie   *Movie
	WatchedMovies   []Movie
}

type PartyMovie struct {
	ID          pgtype.Int8
	Title       pgtype.Text
	PosterURL   pgtype.Text
	WatchStatus pgtype.Text
}

const (
	createPartyQuery       = `INSERT INTO parties (name, short_id) VALUES ($1, $2) returning id_party;`
	createPartyMemberQuery = `INSERT INTO party_members (id_member, id_party, owner) VALUES ($1, $2, true);`
)

var (
	ErrDuplicatePartyName    = errors.New("party name already exists")
	ErrDuplicatePartyShortID = errors.New("party short id already exists")
)

func (p *PGStore) CreateParty(ctx context.Context, idMember int, name, shortID string) (int, error) {
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
				if pgErr.ConstraintName == "idx_parties_short_id" {
					return 0, ErrDuplicatePartyShortID
				}
			}
		}
		return 0, err
	}

	_, err = txn.Exec(ctx, createPartyMemberQuery, idMember, id)
	if err != nil {
		return 0, err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

const getPartiesQueryForMember = `
  select parties.id_party, parties.name from parties
  join party_members on party_members.id_party = parties.id_party
  where party_members.id_member = $1;
`

func (p *PGStore) GetPartiesByMemberID(ctx context.Context, idMember int) ([]Party, error) {
	rows, err := p.db.Query(ctx, getPartiesQueryForMember, idMember)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var parties []Party

	for rows.Next() {
		var party Party
		err := rows.Scan(&party.ID, &party.Name)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

const getPartiesQueryForMovie = `
with filtered_party_movies as(select * from party_movies where id_movie = $1)select parties.id_party, parties.name, movies.id_movie is not null as is_movie
  from parties
  join party_members on party_members.id_party = parties.id_party
  left outer join filtered_party_movies on filtered_party_movies.id_party = parties.id_party 
  left outer join movies on filtered_party_movies.id_movie = movies.id_movie
  where party_members.id_member = $2;
`

func (p *PGStore) GetPartiesByMemberIDForCurrentMovie(ctx context.Context, idMovie int, idMember int) ([]Party, error) {
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

const getPartyByIDQuery = `select id_party, name, short_id from parties where id_party = $1`

func (p *PGStore) GetPartyByID(ctx context.Context, id int) (Party, error) {
	party := Party{}
	err := p.db.QueryRow(ctx, getPartyByIDQuery, id).Scan(&party.ID, &party.Name, &party.ShortID)
	if err != nil {
		return Party{}, err
	}
	return party, nil
}

const getPartyByShortIDQuery = `select id_party, name, short_id from parties where short_id = $1`

func (p *PGStore) GetPartyByShortID(ctx context.Context, shortID string) (Party, error) {
	party := Party{}

	err := p.db.QueryRow(ctx, getPartyByShortIDQuery, shortID).Scan(&party.ID, &party.Name, &party.ShortID)
	if err != nil {
		return Party{}, err
	}

	return party, nil
}

const getPartyByIDWithMoviesQuery = `
  select
    parties.id_party, 
    parties.name,
    parties.short_id,
    movies.id_movie, 
    movies.title,
    movies.poster_url,
    profiles.id_profile,
    profiles.first_name,
    profiles.last_name,
    party_movies.watch_status
  from parties
  left join party_movies on party_movies.id_party = parties.id_party
  left join movies on movies.id_movie = party_movies.id_movie
  left join party_members on party_members.id_party = parties.id_party
  left join profiles on profiles.id_profile = party_members.id_profile
  where parties.id_party = $1;
`

func (p *PGStore) GetPartyByIDWithMovies(ctx context.Context, partyID int) (Party, error) {
	party := Party{UnwatchedMovies: make([]Movie, 0), WatchedMovies: make([]Movie, 0)}
	rows, err := p.db.Query(ctx, getPartyByIDWithMoviesQuery, partyID)
	if err != nil {
		p.logger.Error(err.Error(), "query", getPartyByIDWithMoviesQuery)
		return Party{}, err
	}

	for rows.Next() {
		var (
			pm      PartyMovie
			profile Profile
		)
		err := rows.Scan(&party.ID, &party.Name, &party.ShortID, &pm.ID, &pm.Title, &pm.PosterURL, &profile.ID, &profile.FirstName, &profile.LastName, &pm.WatchStatus)
		if err != nil {
			p.logger.Error(err.Error(), "query", getPartyByIDWithMoviesQuery)
			return Party{}, err
		}

		if pm.ID.Valid {
			movie := Movie{
				ID:          int(pm.ID.Int64),
				Title:       pm.Title.String,
				PosterURL:   pm.PosterURL.String,
				WatchStatus: WatchStatusEnum(pm.WatchStatus.String),
				AddedBy:     profile,
			}
			switch movie.WatchStatus {
			case WatchStatusUnwatched:
				party.UnwatchedMovies = append(party.UnwatchedMovies, movie)
			case WatchStatusSelected:
				party.SelectedMovie = &movie
			case WatchStatusWatched:
				party.WatchedMovies = append(party.WatchedMovies, movie)
			}
		}
	}

	if err := rows.Err(); err != nil {
		p.logger.Error(err.Error(), "query", getPartyByIDWithMoviesQuery)
		return Party{}, err
	}

	p.logger.Info("GetPartyByIDWithMovies", "party", party, "partyID", partyID)

	return party, nil
}

const AddMovietoPartyQuery = `insert into party_movies (id_party, id_movie) values ($1, $2)`

func (p *PGStore) AddMovieToParty(ctx context.Context, idParty, idMovie int) error {
	_, err := p.db.Exec(ctx, AddMovietoPartyQuery, idParty, idMovie)
	if err != nil {
		return err
	}
	return nil
}

const GetPartiesByMemberIDQuery = `
  select parties.id_party, parties.name from parties
  left join party_members on party_members.id_party = parties.id_party
  where party_members.id_member = $1
`

func (pg *PGStore) GetPartiesForMember(ctx context.Context, id int) ([]Party, error) {
	rows, err := pg.db.Query(ctx, GetPartiesByMemberIDQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parties []Party
	for rows.Next() {
		var party Party
		err := rows.Scan(&party.ID, &party.Name)
		if err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, nil
}

const updateWatchStatusQuery = `
  update party_movies
  set watch_status = $1
  where id_party = $2 and id_movie = $3
`

func (pg *PGStore) MarkMovieAsWatched(ctx context.Context, idParty, idMovie int) error {
	pg.logger.Info("MarkMovieAsWatched", "idParty", idParty, "idMovie", idMovie)
	return pg.updateMovieStatusInParty(ctx, idParty, idMovie, WatchStatusWatched)
}

const selectMovieForPartyQuery = `
WITH selected_member_id AS (
  select id_member 
  from party_members 
  where id_party = $1
  order by random()
  limit 1
)
  select id_movie 
  from party_movies 
  where id_party = $2 and id_member = (select id_member from selected_member_id) AND watch_status = 'unwatched'
  order by random()
  limit 1;
`

func (pg *PGStore) SelectMovieForParty(ctx context.Context, idParty int) error {
	var selectedMovieID int
	err := pg.db.QueryRow(ctx, selectMovieForPartyQuery, idParty, idParty).Scan(&selectedMovieID)
	if err != nil {
		return err
	}

	return pg.updateMovieStatusInParty(ctx, idParty, selectedMovieID, WatchStatusSelected)
}

func (pg *PGStore) updateMovieStatusInParty(ctx context.Context, idParty, idMovie int, status WatchStatusEnum) error {
	_, err := pg.db.Exec(ctx, updateWatchStatusQuery, status, idParty, idMovie)
	if err != nil {
		return err
	}
	return nil
}
