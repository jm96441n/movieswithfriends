-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table party_movies (
    id INT GENERATED ALWAYS AS IDENTITY,
    id_party INT,
    id_movie INT,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id),
    CONSTRAINT fk_party_movies_parties FOREIGN KEY(id_party) REFERENCES parties(id_party),
    CONSTRAINT fk_party_movies_movies FOREIGN KEY(id_movie) REFERENCES movies(id_movie)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists party_movies;

