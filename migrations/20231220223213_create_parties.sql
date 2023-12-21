-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table parties (
    id_party INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_party)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists parties;
