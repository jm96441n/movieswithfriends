-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table movies (
    id_movie INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_movie)
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists movies;
