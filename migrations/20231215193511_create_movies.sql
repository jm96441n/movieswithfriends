-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table movies (
    id_movie INT GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(100) NOT NULL,
    poster_url VARCHAR(100) NOT NULL,
    tmdb_id INTEGER NOT NULL,
    overview TEXT NOT NULL,
    tagline TEXT NOT NULL,
    release_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_movie)
);
-- Add an index on the created column.
CREATE INDEX idx_movies_tmdb_id ON movies(tmdb_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists movies;
