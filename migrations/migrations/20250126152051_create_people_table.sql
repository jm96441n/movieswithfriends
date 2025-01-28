-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE people (
    id_people INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ,
    PRIMARY KEY(id_actor)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS people;
-- +goose StatementEnd
