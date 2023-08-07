-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table party_movies (
    id_party INT,
    id_movie INT,
    CONSTRAINT fk_movie FOREIGN KEY(id_movie) REFERENCES movies(id_movie),
    CONSTRAINT fk_party FOREIGN KEY(id_party) REFERENCES parties(id_party)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists party_movies;
