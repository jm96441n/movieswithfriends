-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table profile_parties add CONSTRAINT unique_profile_per_party UNIQUE(id_party, id_profile);
alter table party_movies add CONSTRAINT unique_movie_per_party UNIQUE(id_party, id_movie);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table party_movies drop CONSTRAINT unique_movie_per_party;
alter table profile_parties drop CONSTRAINT unique_profile_per_party;
-- +goose StatementEnd
