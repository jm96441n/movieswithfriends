-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
create table profile_parties (
    id_party INT,
    id_profile INT,
    CONSTRAINT fk_profile FOREIGN KEY(id_profile) REFERENCES profiles(id_profile),
    CONSTRAINT fk_party FOREIGN KEY(id_party) REFERENCES parties(id_party)
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
drop table if exists party_movies;
