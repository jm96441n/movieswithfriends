package store

import "context"

type Party struct {
	Name            string
	ID              int
	MovieAdded      bool
	UnwatchedMovies []Movie
	SelectedMovie   *Movie
	WatchedMovies   []Movie
}

const getPartiesQuery = `
with filtered_party_movies as(select * from party_movies where id_movie = $1)select parties.id_party, parties.name, movies.id_movie is not null as is_movie
  from parties
  left outer join filtered_party_movies on filtered_party_movies.id_party = parties.id_party left outer join movies on filtered_party_movies.id_movie = movies.id_movie;
`

func (p *PGStore) GetParties(ctx context.Context, idMovie int) ([]Party, error) {
	rows, err := p.db.Query(ctx, getPartiesQuery, idMovie)
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

const getPartyByIDQuery = `select id_party, name from parties where id_party = $1`

func (p *PGStore) GetPartyByID(ctx context.Context, id int) (Party, error) {
	party := Party{}
	err := p.db.QueryRow(ctx, getPartyByIDQuery, id).Scan(&party.ID, &party.Name)
	if err != nil {
		return Party{}, err
	}
	return party, nil
}

const getPartyByIDWithMoviesQuery = `
  select
    parties.id_party, 
    parties.name,
    movies.id_movie, 
    movies.title,
    movies.poster_url,
    profiles.id_profile,
    profiles.first_name,
    profiles.last_name,
    party_movies.watch_status
  from parties
  join party_movies on party_movies.id_party = parties.id_party
  join movies on movies.id_movie = party_movies.id_movie
  join profiles on profiles.id_profile = party_movies.id_profile
  where parties.id_party = $1;
`

func (p *PGStore) GetPartyByIDWithMovies(ctx context.Context, id int) (Party, error) {
	party := Party{UnwatchedMovies: make([]Movie, 0), WatchedMovies: make([]Movie, 0)}
	rows, err := p.db.Query(ctx, getPartyByIDWithMoviesQuery, id)
	if err != nil {
		p.logger.Error(err.Error(), "query", getPartyByIDWithMoviesQuery)
		return Party{}, err
	}

	for rows.Next() {
		var movie Movie
		err := rows.Scan(&party.ID, &party.Name, &movie.ID, &movie.Title, &movie.PosterURL, &movie.AddedBy.ID, &movie.AddedBy.FirstName, &movie.AddedBy.LastName, &movie.WatchStatus)
		if err != nil {
			p.logger.Error(err.Error(), "query", getPartyByIDWithMoviesQuery)
			return Party{}, err
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

const GetPartiesByProfileIDQuery = `
  select parties.id_party, parties.name from parties
  left join profile_parties on profile_parties.id_party = parties.id_party
  where profile_parties.id_profile = $1
`

func (pg *PGStore) GetPartiesForProfile(ctx context.Context, id int) ([]Party, error) {
	rows, err := pg.db.Query(ctx, GetPartiesByProfileIDQuery, id)
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
