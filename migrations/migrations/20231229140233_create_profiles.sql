-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

create table profiles (
    id_profile INT GENERATED ALWAYS AS IDENTITY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ,
    PRIMARY KEY(id_profile)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

drop table if exists profiles;
-- +goose StatementEnd
