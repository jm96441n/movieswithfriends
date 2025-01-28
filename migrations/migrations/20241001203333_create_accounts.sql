-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table accounts (
    id_account INT GENERATED ALWAYS AS IDENTITY,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_account)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists accounts;
